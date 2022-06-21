package app

import (
	"context"
	"fmt"
	"k8s-installer/pkg/util/fileutils"
	"net"
	"reflect"
	"strings"
	"sync"
	"time"

	"k8s-installer/node/reportor"
	runtimeCache "k8s-installer/pkg/cache"
	"k8s-installer/pkg/cmd/client"
	clientConfig "k8s-installer/pkg/config/client"
	cfg "k8s-installer/pkg/config/viper"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/message_queue/nats"
	"k8s-installer/pkg/network"
	"k8s-installer/pkg/node_identity"
	"k8s-installer/pkg/server/version"
	"k8s-installer/pkg/task_breaker"
	"k8s-installer/schema"

	natServer "github.com/nats-io/nats-server/v2/server"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var currentConfig clientConfig.Config
var ctx context.Context
var cancellation context.CancelFunc
var wg sync.WaitGroup

func NewServerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "k8s-install-client",
		Short: "A client intend to dynamically install kubernetes clusters",
		Long:  "A client intend to dynamically install kubernetes clusters \n It listens to the message queue and do stuff against message subject",
		Run: func(cmd *cobra.Command, args []string) {
			commandArgsValidation()
			startClient(cmd, args)
		},
	}
	AddFlags(cmd.PersistentFlags())

	// add a sub command version to show version by command
	// client version
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
	return cmd
}

func LoadConfigFileIfFound() {
	currentConfig = clientConfig.DefaultConfig()
	clientConfig.DefaultConfigLookup()
	log.DefaultLoginSetting(network.GetDefaultIP, currentConfig.Log.LogRetentionPeriod)
	if err := cfg.ParseViperConfig(&currentConfig, cfg.DecodeOptionYaml()); err != nil {
		log.Debugf("No configuration file config-client.(yml|yaml)\"  \" found in any of location %s", "$HOME/.k8s-installer /etc/k8s-installer")
		log.Debug("Using default setting")
	} else {
		log.Debug("Load config from file done!")
	}
	generatorNodeIdIfNotSet(&currentConfig)
	checkConfig()
	cache := runtimeCache.InitCacheRuntime(currentConfig.Cache, currentConfig.Etcd, task_breaker.SetOperationStep)
	if err := cache.SaveOrUpdateClientRuntimeConfig(currentConfig.ClientId, currentConfig); err != nil {
		log.Fatalf("Failed to save node config with id %s to cache type %s", currentConfig.ClientId, reflect.TypeOf(cache).String())
	}
	/*	// after config set load cache data from database
		if err := cache.SyncFromDatabase(); err != nil {
			log.Fatalf("Cache data initialization failed due to error %s ... aborting", err)
		}*/
	log.SetLogLevel(currentConfig.Log.LogLevel)
}

func generatorMessageQueueConfig() *natServer.Options {
	return &natServer.Options{
		Port:     currentConfig.MessageQueue.Port,
		Username: currentConfig.MessageQueue.UserName,
		Password: currentConfig.MessageQueue.Password,
	}
}

func startClient(cmd *cobra.Command, args []string) {
	log.Debugf("Waiting for message queue server %s:%d online", currentConfig.MessageQueue.ServerAddress, currentConfig.MessageQueue.Port)
	targetMQServer := strings.Split(currentConfig.MessageQueue.ServerAddress, ",")
	index := 0
	for {
		log.Debugf("Try to reach mq server %d with ip %s", index, targetMQServer[index])
		if err := network.CheckIpIsReachable(targetMQServer[index], currentConfig.MessageQueue.Port, "tcp", 3*time.Second); err != nil {
			log.Warnf("Failed to reach mq server %d with ip %s", index, targetMQServer[index])
			time.Sleep(3 * time.Second)
		} else {
			log.Debugf("Message queue server %s:%d established", targetMQServer[index], currentConfig.MessageQueue.Port)
			break
		}
		if index < len(targetMQServer)-1 {
			index += 1
		} else {
			log.Warnf("All mq server is not reachable now back to first mq server")
			index = 0
		}
	}

	// attempt to turn off selinux and firewalld
	// or server will never be able to reach client with signal port
	if clientOSFamily, err := client.GetOSFamily("client-init"); err != nil {
		log.Fatalf("Failed to create os family object due to error %s", err)
	} else {
		if _, err := clientOSFamily.GetOSVersion().BasicNodeSetup(schema.TaskBasicConfig{Action: constants.ActionCreate}, schema.Cluster{}); err != nil {
			log.Fatalf("Failed to executor node preparation script due to error %s", err)
		}
	}

	ctx, cancellation = context.WithCancel(context.Background())
	go network.OpenPortAndListen(net.IPv4(0, 0, 0, 0), currentConfig.SignalPort, "", ctx.Done())
	format := "%s.%s"
	addr := network.GetDefaultIP(true)
	subject := fmt.Sprintf(format, addr.To4().String(), currentConfig.MessageQueue.SubjectSuffix)
	if err := nats.ListenMQUp(subject, currentConfig.ClientId, currentConfig.MessageQueue, client.MessageHandler); err != nil {
		log.Fatalf("Failed to setup mq connection due to error %v", err)
	}
	// report node stat to server
	reportor.NodeStatReport(ctx.Done())

	// wait main goroutine forever
	select {}
}

func generatorNodeIdIfNotSet(config *clientConfig.Config) {
	// generator an id for node if ServerId is not set yet
	if config.ClientId == "" {
		//ipv4GW := network.GetDefaultIP(true)
		// so far we use ipv4 gateway as seed make sure with save ip will always get same id
		//config.ClientId = node_identity.GeneratorNodeID(ipv4GW.String(), "client", "")
		config.ClientId = node_identity.GeneratorNodeIDWithUUID("client", "")
		data, errParse := yaml.Marshal(config)
		if errParse != nil {
			log.Fatalf("Failed to save client id to local config due to error %s", errParse.Error())
		} else {
			if err := cfg.WriteViperConfig(data); err != nil {
				log.Fatalf("Failed to save client id to local config due to error %s", err.Error())
			}
		}

	}
}

func checkConfig() {
	if currentConfig.ProxyIpv4CIDR != "" {
		ip, _, err := net.ParseCIDR(currentConfig.ProxyIpv4CIDR)
		if err != nil {
			log.Fatalf("%s is not a valid ip address. Config error abort...", currentConfig.ProxyIpv4CIDR)
		}
		currentConfig.ProxyIpv4CIDR = ip.To4().String()
	}

	if currentConfig.Region == "" {
		log.Warnf("Region is not proper set. Set to default region \"not-set\"")
		currentConfig.Region = "not-set"
	}

	// create YamlDataDir if it does not exist
	fileutils.CreateDirectory(currentConfig.YamlDataDir)
}
