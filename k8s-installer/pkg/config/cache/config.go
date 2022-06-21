package cache

type CacheConfig struct {
	CacheRuntime string `yaml:"cache-runtime"`
}

func DefaultConfig() CacheConfig {
	return CacheConfig{
		CacheRuntime: "no-cache",
	}
}
