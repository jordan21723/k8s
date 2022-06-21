package coredns

type Config struct {
	Subject   string `yaml:"subject"`
	NotifyDNS string `yaml:"notify-mode" description:"how to notify dns either 'now' or 'never'"`
}

func DefaultConfig() Config {
	return Config{
		Subject:   "caas-dns",
		NotifyDNS: "never",
	}
}
