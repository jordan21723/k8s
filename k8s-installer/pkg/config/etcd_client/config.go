package etcd_client

type EtcdConfig struct {
	AuthMode   string `yaml:"etcd-auth-mode"`
	EndPoints  string `yaml:"etcd-endpoints"`
	Username   string `yaml:"etcd-username"`
	Password   string `yaml:"etcd-password"`
	CaCertFile string `yaml:"etcd-ca-file-path"`
	CertFile   string `yaml:"etcd-cert-file-path"`
	KeyFile    string `yaml:"etcd-key-file-path"`
}

func DefaultConfig() EtcdConfig {
	return EtcdConfig{
		AuthMode:   "tls",
		EndPoints:  "https://localhost:2379 | https://localhost:2379 | https://localhost:2379",
		CaCertFile: "/etcd/etcd-client-ca.crt",
		CertFile:   "/etcd/etcd-client.crt",
		KeyFile:    "/etcd/etcd-client.key",
	}
}
