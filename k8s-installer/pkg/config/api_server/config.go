package api_server

type ApiConfig struct {
	ApiVipAddress          string `yaml:"api-vip-address"`
	ApiPort                uint32 `yaml:"api-port"`
	ResourceServerPort     uint32 `yaml:"resource-server-port"`
	ResourceServerFilePath string `yaml:"resource-server-filepath"`
	ResourceServerCIDR     string `yaml:"resource-server-cidr"`
	EnableTLS              bool   `yaml:"enable-tls"`
	JWTSignString          string `yaml:"jwt-sign-string"`
	JWTSignMethod          string `yaml:"jwt-sign-method"`
	JWTExpiredAfterHours   uint32 `yaml:"jwt-expired-after-hours"`
}
