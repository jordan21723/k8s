package app

import "github.com/spf13/pflag"

func AddFlags(flags *pflag.FlagSet) {
	flags.StringVarP(&currentConfig.MessageQueue.ServerListeningOnIp, "mq-server-listening-ip", "", currentConfig.MessageQueue.ServerListeningOnIp, "message queue listening on ip default 0.0.0.0")
	flags.IntVarP(&currentConfig.MessageQueue.Port, "mq-server-port", "", currentConfig.MessageQueue.Port, "message queue server listening on port. Default 9889")
	flags.IntVarP(&currentConfig.MessageQueue.ClusterCommunicationPort, "mq-cluster-port", "", currentConfig.MessageQueue.ClusterCommunicationPort, "message queue cluster Communication port. Default 9890")
	flags.StringVarP(&currentConfig.MessageQueue.AuthMode, "mq-auth", "a", currentConfig.MessageQueue.AuthMode, "message queue auth mode either basic or tls")
	flags.StringVarP(&currentConfig.MessageQueue.UserName, "mq-username", "u", currentConfig.MessageQueue.UserName, "message queue user name")
	flags.StringVarP(&currentConfig.MessageQueue.Password, "mq-password", "p", currentConfig.MessageQueue.Password, "message queue password")
	flags.DurationVarP(&currentConfig.MessageQueue.ReconnectWait, "mq-reconnect-wait", "", currentConfig.MessageQueue.ReconnectWait, "message queue reconnect interval")
	flags.IntVarP(&currentConfig.MessageQueue.MaxReconnect, "mq-max-reconnect", "", currentConfig.MessageQueue.MaxReconnect, ""+
		"MaxReconnect sets the number of reconnect attempts that will be tried before giving up. "+
		"If negative, then it will never give up trying to reconnect.")
	flags.StringVarP(&currentConfig.MessageQueue.TlsCaPath, "ca-cert", "s", currentConfig.MessageQueue.TlsCaPath, "message queue ca cert file path")
	flags.StringVarP(&currentConfig.MessageQueue.TlsCertPath, "cert", "c", currentConfig.MessageQueue.TlsCertPath, "message queue cert file path")
	flags.StringVarP(&currentConfig.MessageQueue.TlsCaPath, "key", "k", currentConfig.MessageQueue.TlsCaPath, "message queue key file path")
	flags.StringVarP(&currentConfig.MessageQueue.QueueGroupName, "queue-group-name", "", currentConfig.MessageQueue.QueueGroupName, "queue group name used to distribute message to one of a group listener member")
	flags.StringVarP(&currentConfig.MessageQueue.ServerNodeStatusReportInSubject, "queue-node-report-subject", "", currentConfig.MessageQueue.ServerNodeStatusReportInSubject, "subject channel for node status report message queue channel")
	flags.Uint32VarP(&currentConfig.Log.LogLevel, "log-level", "l", currentConfig.Log.LogLevel, "log level max 6")
	flags.Uint32VarP(&currentConfig.ApiServer.ApiPort, "api-server-port", "z", currentConfig.ApiServer.ApiPort, "Api server port number")
	flags.StringVarP(&currentConfig.Cache.CacheRuntime, "cache-runtime", "", currentConfig.Cache.CacheRuntime, "cache runtime setting either go-cache or local-ram or group-cache or no-cache default local-ram")
	flags.StringVarP(&currentConfig.Etcd.AuthMode, "etcd-auth-mode", "", currentConfig.Etcd.AuthMode, "etcd auth mode either basic or tls")
	flags.StringVarP(&currentConfig.Etcd.EndPoints, "etcd-endpoints", "", currentConfig.Etcd.EndPoints, "etcd endpoints such as http(s)://10.0.0.100:2379 | http(s)://10.0.0.101:2379 | http(s)://10.0.0.102:2379")
	flags.StringVarP(&currentConfig.Etcd.Username, "etcd-username", "", currentConfig.Etcd.Username, "etcd username only needed when auth mode set to basic")
	flags.StringVarP(&currentConfig.Etcd.Password, "etcd-password", "", currentConfig.Etcd.Password, "etcd password only needed when auth mode set to basic")
	flags.StringVarP(&currentConfig.Etcd.CaCertFile, "etcd-ca-file-path", "", currentConfig.Etcd.CaCertFile, "etcd ca file path only needed when auth mode set to tls")
	flags.StringVarP(&currentConfig.Etcd.CertFile, "etcd-cert-file-path", "", currentConfig.Etcd.CertFile, "etcd certification file path only needed when auth mode set to tls")
	flags.StringVarP(&currentConfig.Etcd.KeyFile, "etcd-key-file-path", "", currentConfig.Etcd.KeyFile, "etcd key file path only needed when auth mode set to tls")
	flags.IntVarP(&currentConfig.SignalPort, "client-signal-port", "", currentConfig.SignalPort, "set client signal port in order to help server can test it when needed")
}
