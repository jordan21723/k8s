package log

type LogConfig struct {
	LogLevel           uint32 `yaml:"log-level"`
	LogRetentionPeriod int    `yaml:"log-retention-period"`
}

func DefaultConfig() LogConfig {
	return LogConfig{
		LogLevel:           6,
		LogRetentionPeriod: 30,
	}
}
