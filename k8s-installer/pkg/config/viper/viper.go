package viper

import (
	"bytes"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func ParseViperConfig(config interface{}, options ...viper.DecoderConfigOption) (err error) {
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	return viper.Unmarshal(&config, options...)
}

func WriteViperConfig(content []byte) (err error) {
	err = viper.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		return
	}
	return viper.WriteConfig()
}

func DecodeOptionYaml() func(config *mapstructure.DecoderConfig) {
	return func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	}
}
