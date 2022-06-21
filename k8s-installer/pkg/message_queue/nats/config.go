package nats

import "time"

const (
	AUTHMODE_BASIC = "basic"
	AUTHMODE_TLS   = "tls"
)

/*
ServerAddress: when use a client , we need this var to connect to server. Example: "nats://10.0.0.10:4222,nats://10.0.0.11:4222,nats://10.0.0.12:4222" or single server "nats://10.0.0.10:4222"
QueueGroupName: is for multiple servers scenario,message should forward only one server to handler it
*/

type MessageQueueConfig struct {
	ServerAddress                   string `yaml:"message-queue-server-address"`
	ServerListeningOnIp             string `yaml:"message-queue-server-listening-on"`
	ClusterLeaderIp                 string `yaml:"message-queue-leader"`
	ClusterCommunicationPort        int    `yaml:"cluster-port"`
	SubjectSuffix                   string `yaml:"message-queue-subject-Suffix"`
	AuthMode                        string `yaml:"auth-mode"`
	Port                            int    `yaml:"port"`
	UserName                        string `yaml:"username"`
	Password                        string `yaml:"password"`
	TlsCertPath                     string `yaml:"cert-path"`
	TlsKeyPath                      string `yaml:"key-path"`
	TlsCaPath                       string `yaml:"ca-cert-path"`
	QueueGroupName                  string `yaml:"message-queue-group-name"`
	ServerNodeStatusReportInSubject string `yaml:"message-queue-node-status-report-in-subject"`
	TimeOutSeconds                  int    `yaml:"time-out-seconds"`

	// ReconnectWait sets the time to backoff after attempting a reconnect
	// to a server that we were already connected to previously.
	ReconnectWait time.Duration `yaml:"reconnectWait"`

	// MaxReconnect sets the number of reconnect attempts that will be
	// tried before giving up. If negative, then it will never give up
	// trying to reconnect.
	MaxReconnect       int `yaml:"maxReconnect"`
	PintInterval       int `yaml:"pint_interval"`
	MaxPingOutStanding int `yaml:"max_ping_out_standing"`
}

func DefaultConfig() MessageQueueConfig {
	return MessageQueueConfig{
		ServerListeningOnIp:             "0.0.0.0",
		ServerAddress:                   "127.0.0.1",
		ClusterLeaderIp:                 "127.0.0.1", // cluster seed ip = first instance of cluster
		ClusterCommunicationPort:        9890,
		AuthMode:                        AUTHMODE_BASIC,
		Port:                            9889,
		UserName:                        "username",
		Password:                        "password",
		TlsCertPath:                     "",
		TlsKeyPath:                      "",
		TlsCaPath:                       "",
		SubjectSuffix:                   "k8s-install",
		QueueGroupName:                  "caas-node-status-report-queue",
		ServerNodeStatusReportInSubject: "caas-node-status-report-subject",
		TimeOutSeconds:                  10, // require reply message time out
		ReconnectWait:                   1 * time.Second,
		MaxReconnect:                    600,
		PintInterval:                    20,
		MaxPingOutStanding:              5,
	}
}
