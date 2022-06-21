package app

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	depRegistry "k8s-installer/pkg/dep_registry"
	"k8s-installer/pkg/server/version"

	bd "k8s-installer/pkg/block_device"
	"k8s-installer/pkg/server/client_liveness"

	"k8s-installer/pkg/task_breaker"

	apiServer "k8s-installer/internal/apiserver"
	runtimeCache "k8s-installer/pkg/cache"
	"k8s-installer/pkg/config/server"
	cfg "k8s-installer/pkg/config/viper"
	"k8s-installer/pkg/control_manager"
	jwtPkg "k8s-installer/pkg/jwt"
	"k8s-installer/pkg/log"
	messageQueue "k8s-installer/pkg/message_queue/nats"
	"k8s-installer/pkg/network"
	"k8s-installer/pkg/node_identity"

	"github.com/google/uuid"
	natServer "github.com/nats-io/nats-server/v2/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var currentConfig server.Config
var ctx context.Context
var cancellation context.CancelFunc
var wg sync.WaitGroup

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s-install-server",
		Short: "A server intend to dynamically install kubernetes clusters",
		Long:  "A server intend to dynamically install kubernetes clusters \n It contains three main components \n 1. Message Queue Server \n 2. An restful style api server \n 3. CLI itself",
		Run: func(cmd *cobra.Command, args []string) {
			LoadConfigFileIfFound()
			if strings.ToLower(currentConfig.MessageQueue.AuthMode) != "basic" && strings.ToLower(currentConfig.MessageQueue.AuthMode) != "tls" {
				log.Fatalf("Message queue auth mode %s is not valid only basic or tls is support", currentConfig.MessageQueue.AuthMode)
			}
			startServer(cmd, args)
		},
	}
	AddFlags(cmd.PersistentFlags())

	// add a sub command version to show version by command
	// server version
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "show version",
		Long:  `show current binary version`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print(version.Version)
		},
	}

	cmd.AddCommand(versionCmd)

	// add a sub command to generator all dependencies json file
	genDepsCmd := &cobra.Command{
		Use:   "genDeps",
		Short: "generator dependencies list",
		Long:  `generator a dependencies file on /tmp/k8s-installer-deps.json`,
		Args:  cobra.MaximumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			depRegistry.GenDepsFile()
		},
	}

	cmd.AddCommand(genDepsCmd)

	return cmd
}

func LoadConfigFileIfFound() {
	currentConfig = server.DefaultConfig()
	server.DefaultConfigLookup()
	log.DefaultLoginSetting(network.GetDefaultIP, currentConfig.Log.LogRetentionPeriod)
	if err := cfg.ParseViperConfig(&currentConfig, cfg.DecodeOptionYaml()); err != nil {
		log.Debugf("No configuration file config-server.(yml|yaml)\"  \" found in any of location %s", "$HOME/.k8s-installer /etc/k8s-installer")
		log.Debug("Using default setting")
	} else {
		log.Debug("Load config from file done!")
	}

	if !currentConfig.IsTestNode {
		checkProductionEnvironment()
	}

	generatorNodeIdIfNotSet(&currentConfig)
	// get runtime cache
	cache := runtimeCache.InitCacheRuntime(currentConfig.Cache, currentConfig.Etcd, task_breaker.SetOperationStep)

	// check server config stop if error
	serverConfigCheck()
	if err := cache.SaveOrUpdateServerRuntimeConfig(currentConfig.ServerId, currentConfig); err != nil {
		log.Fatalf("Failed to save node config with id %s to cache type %s ... aborting", currentConfig.ServerId, reflect.TypeOf(cache).String())
	}
	// after config set load cache data from database
	if err := cache.SyncFromDatabase(); err != nil {
		log.Fatalf("Cache data initialization failed due to error %s ... aborting", err)
	}

	// set log level to the one in config
	log.SetLogLevel(currentConfig.Log.LogLevel)
}

func startServer(cmd *cobra.Command, args []string) {
	ctx, cancellation = context.WithCancel(context.Background())
	// start message queue server
	go messageQueue.StartMessageQueue(generatorMessageQueueConfig(), ctx.Done())

	// node agent liveness detector
	if currentConfig.EnableAgentLivenessDetector {
		go client_liveness.ClientAgentDaemon(ctx.Done())
	}

	// run resource server http restful api
	go apiServer.Start(currentConfig.ApiServer, currentConfig.ApiServer.ResourceServerPort, apiServer.ResourceHandler(currentConfig.ApiServer.ResourceServerFilePath), ctx.Done(), nil)
	// run api server with go-restful
	go apiServer.Start(currentConfig.ApiServer, currentConfig.ApiServer.ApiPort, apiServer.CreateGoRestful(), ctx.Done(), nil)

	// start node report in queue
	time.Sleep(1 * time.Second)
	if err := messageQueue.ListenQueueGroup(currentConfig.MessageQueue.ServerNodeStatusReportInSubject,
		uuid.New().String(),
		currentConfig.MessageQueue.QueueGroupName,
		currentConfig.MessageQueue,
		control_manager.NodeStateReportInHandler); err != nil {
		log.Fatalf("Failed to setup mq connection due to error %v", err)
	}

	// wait main goroutine forever
	select {}
}

func generatorMessageQueueConfig() *natServer.Options {
	return &natServer.Options{
		Port:     currentConfig.MessageQueue.Port,
		Username: currentConfig.MessageQueue.UserName,
		Password: currentConfig.MessageQueue.Password,
		//Host:     currentConfig.MessageQueue.ServerListeningOnIp,
		Cluster: natServer.ClusterOpts{
			Host: currentConfig.MessageQueue.ServerListeningOnIp,
			Port: currentConfig.MessageQueue.ClusterCommunicationPort,
			//Username: currentConfig.MessageQueue.UserName,
			//Password: currentConfig.MessageQueue.UserName,
		},
		// we use one of master ip as the leader
		// all of masters have to set message-queue-leader to this master ip will do the trick e.g. message-queue-leader: 10.0.0.100
		// when use message queue cluster client can set multiple server addresses split by command
		// e.g. 10.0.0.100:2222,10.0.0.101:2222,10.0.0.102:2222
		// won`t abort program even the ip:port not open yet
		Routes: natServer.RoutesFromStr(fmt.Sprintf("nats://%s:%d", currentConfig.MessageQueue.ClusterLeaderIp, currentConfig.MessageQueue.ClusterCommunicationPort)),
	}
}

func generatorNodeIdIfNotSet(config *server.Config) {
	// generator an id for node if ServerId is not set yet
	if config.ServerId == "" {
		ipv4GW := network.GetDefaultIP(true)
		// so far we use ipv4 gateway as seed make sure with same ip will always get same id
		config.ServerId = node_identity.GeneratorNodeID(ipv4GW.String(), "server", "")
		data, errParse := yaml.Marshal(config)
		if errParse != nil {
			log.Fatalf("Failed to save server id to local config due to error %s", errParse.Error())
		} else {
			if err := cfg.WriteViperConfig(data); err != nil {
				log.Fatalf("Failed to save server id to local config due to error %s", err.Error())
			}
		}
	}
}

func serverConfigCheck() {
	//resourceServerCIDR := currentConfig.ApiServer.ResourceServerCIDR
	if currentConfig.ApiServer.ResourceServerCIDR == "" {
		// if resource cidr is not set use default ipv4
		currentConfig.ApiServer.ResourceServerCIDR = network.GetDefaultIP(true).String()
	} else {
		ip, ipNet, err := net.ParseCIDR(currentConfig.ApiServer.ResourceServerCIDR)
		if err != nil {
			log.Fatalf("%s is not a valid ip address. Config error abort...", currentConfig.ApiServer.ResourceServerCIDR)
		}
		if _, err := network.GetNicWithCIDR(ipNet.String()); err != nil {
			log.Warnf("Unable to found cidr %s on localhost route. It may result in client unable to find resource server error if no proper network is pre-config", ipNet.String())
		}
		// remove mask
		currentConfig.ApiServer.ResourceServerCIDR = ip.String()
	}

	if _, err := jwtPkg.GetJWTSignMethod(currentConfig.ApiServer.JWTSignMethod); err != nil {
		log.Warnf("JWT sign method %s is not valid fall back to default HS256", currentConfig.ApiServer.JWTSignMethod)
		currentConfig.ApiServer.JWTSignMethod = "HS256"
	}

	if currentConfig.ApiServer.JWTSignString == "" {
		log.Warnf("No jwt sign string is set use default sign string azhzLWluc3RhbGxlcgo=")
		currentConfig.ApiServer.JWTSignString = "azhzLWluc3RhbGxlcgo="
	}
}

func checkProductionEnvironment() {
	var errorList []string
	// check /var/lib/etcd/
	if result, err := bd.CheckDirWithMountSize("/var/lib/etcd/", currentConfig.EtcdDevMinDiskSize); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/etcd/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.EtcdDevMinDiskSize+1))
	} else if !result {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/etcd/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.EtcdDevMinDiskSize+1))
	}

	// check /var/lib/docker/
	if result, err := bd.CheckDirWithMountSize("/var/lib/docker/", currentConfig.DockerDevMinDiskSize); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/docker/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.DockerDevMinDiskSize+1))
	} else if !result {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/docker/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.DockerDevMinDiskSize+1))
	}

	// check /var/lib/registry
	if result, err := bd.CheckDirWithMountSize("/var/lib/registry/", currentConfig.RegistryDevMinDiskSize); err != nil {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/registry/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.RegistryDevMinDiskSize+1))
	} else if !result {
		errorList = append(errorList, fmt.Sprintf("Failed to check dir /var/lib/registry/ mount point. Does it proper mounted and has %dG disk space available?", currentConfig.RegistryDevMinDiskSize+1))
	}

	if len(errorList) > 0 {
		log.Error("Server production environment check failed please see following for detail")
		for _, msg := range errorList {
			log.Error(msg)
		}
		log.Fatal("Aborting...")
	}
}
