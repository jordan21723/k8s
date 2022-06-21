package nats

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"k8s-installer/pkg/log"

	natServer "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const (
	MQ_AUTHENTICATION_TYPE_BASIC = "basic"
	MQ_AUTHENTICATION_TYPE_TLS   = "tls"
)

var (
	nc *nats.Conn
)

func StartMessageQueue(config *natServer.Options, stopChan <-chan struct{}) {
	server, err := natServer.NewServer(config)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Start NATS message queue server and listening on port %s:%d", config.Host, config.Port)
	log.Debugf("NATS message queue cluster port %s:%d", config.Cluster.Host, config.Cluster.Port)
	go server.Start()
	<-stopChan
	log.Debug("Stop signal received. Nats message queue server shutting down !!!")
	server.Shutdown()
}

func ListenMQUp(subject, connectionID string, mq MessageQueueConfig, handler func(msg *nats.Msg)) (err error) {
	opts := []nats.Option{nats.Name(connectionID)}

	opts = setupConnOptions(opts, mq)
	nc, err = nats.Connect(mq.getConnectionString(), opts...)
	if err != nil {
		return errors.New("Error creating client: " + err.Error())
	}
	defer func() {
		if err != nil {
			nc.Close()
		}
	}()
	log.Debugf("Listening on subscribe %s", subject)
	if _, err = nc.Subscribe(subject, handler); err != nil {
		return
	}
	err = nc.Flush()
	return
}

func ListenQueueGroup(subject, clientId, queueName string, mq MessageQueueConfig, handler func(msg *nats.Msg)) (err error) {
	opts := []nats.Option{nats.Name(clientId)}

	opts = setupConnOptions(opts, mq)
	nc, err = nats.Connect(mq.getConnectionString(), opts...)
	if err != nil {
		return errors.New("Error creating client: " + err.Error())
	}
	log.Debugf("Listening on subscribe %s with load balance style queue group", subject)
	defer func() {
		if err != nil {
			nc.Close()
		}
	}()
	if _, err = nc.QueueSubscribe(subject, queueName, handler); err != nil {
		return
	}
	return
}

func SendingMessage(subscribe, data, clientId string) error {
	log.Debugf("Sending message %s to channel %s message client id %s", subscribe, data, clientId)
	return nc.Publish(subscribe, []byte(data))
}

func SendingMessageWithReply(subscribe, senderNodeId, toNodeId, nodeStepId string, mq MessageQueueConfig, secTimeout time.Duration, data []byte, replyHandler func(msg *nats.Msg) error, timeOutHandler func(nodeId string, nodeStepId string)) error {
	log.Debugf("Sending message %s to channel %s message client id %s", data, subscribe, toNodeId)
	msg, err := nc.Request(subscribe, data, secTimeout)
	if err != nil {
		log.Error(err)
		if errors.Is(err, nats.ErrTimeout) && timeOutHandler != nil {
			timeOutHandler(toNodeId, nodeStepId)
		}
		return err
	}
	return replyHandler(msg)
}

func SendingMessageAndWaitReply(subscribe, clientId, nodeId, nodeStepId string, mq MessageQueueConfig, secTimeout time.Duration, data []byte) ([]byte, error) {
	log.Debugf("Sending message %s to channel %s message client id %s", data, subscribe, clientId)
	msg, err := nc.Request(subscribe, data, secTimeout)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return msg.Data, nil
}

func (nat MessageQueueConfig) getConnectionString() string {
	switch nat.AuthMode {
	case MQ_AUTHENTICATION_TYPE_BASIC:
		serverAddress := nat.combineServerAddress()
		// we should split connection string first
		// then combine things like [user]:[password]@[address]
		if strings.Contains(serverAddress, ",") {
			var result []string
			addresses := strings.Split(serverAddress, ",")
			for _, address := range addresses {
				result = append(result, nat.combineUsernameAndPassword()+"@"+address)
			}
			return strings.Join(result, ",")
		} else {
			return nat.combineUsernameAndPassword() + "@" + nat.combineServerAddress()
		}
	default:
		return nat.combineServerAddress()
	}
}

func (conf MessageQueueConfig) combineUsernameAndPassword() string {
	return conf.UserName + ":" + conf.Password
}

func (conf MessageQueueConfig) combineServerAddress() string {
	port := strconv.Itoa(conf.Port)
	if strings.Contains(conf.ServerAddress, ",") {
		var result []string
		addresses := strings.Split(conf.ServerAddress, ",")
		for _, address := range addresses {
			result = append(result, address+":"+port)
		}
		return strings.Join(result, ",")
	} else {
		return conf.ServerAddress + ":" + port
	}
}

func setupConnOptions(opts []nats.Option, mq MessageQueueConfig) []nats.Option {
	opts = append(opts, nats.ReconnectWait(mq.ReconnectWait))
	opts = append(opts, nats.MaxReconnects(mq.MaxReconnect))
	opts = append(opts, nats.PingInterval(time.Duration(mq.PintInterval)*time.Second))
	opts = append(opts, nats.MaxPingsOutstanding(mq.MaxPingOutStanding))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		if err != nil {
			log.Debugf("Disconnected due to:%s, will attempt reconnects for %d times", err, mq.MaxReconnect)
		}
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Debugf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	switch mq.AuthMode {
	case strings.ToLower(MQ_AUTHENTICATION_TYPE_BASIC):
		opts = append(opts, nats.UserInfo(mq.UserName, mq.Password))
	case strings.ToLower(MQ_AUTHENTICATION_TYPE_TLS):
		opts = append(opts, nats.ClientCert(mq.TlsCertPath, mq.TlsKeyPath))
		opts = append(opts, nats.RootCAs(mq.TlsCaPath))
	}
	return opts
}
