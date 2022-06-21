package server

import (
	"testing"

	"gopkg.in/yaml.v2"
	app "k8s-installer/pkg/config/server"
	cfg "k8s-installer/pkg/config/viper"
	testUtil "k8s-installer/pkg/util/test"
)

var config app.Config

func TestConfig(t *testing.T) {
	t.Run("test file config", testConfigLookup)
	t.Run("test write to config", testUpdateConfigFile)
}

func testConfigLookup(t *testing.T) {
	app.DefaultConfigLookup()
	err := cfg.ParseViperConfig(&config, cfg.DecodeOptionYaml())
	if err != nil {
		t.Errorf("Test failed due to %s:", err.Error())
	}
	testUtil.IntIsExpected(t, "MessageQueue.Port", []int{9889}, []int{config.MessageQueue.Port})
}

func testUpdateConfigFile(t *testing.T) {
	config.MessageQueue.Port = 9999
	if data, err := yaml.Marshal(&config); err != nil {
		t.Errorf("Test failed due to %s:", err.Error())
	} else {
		if err := cfg.WriteViperConfig(data); err != nil {
			t.Errorf("Test failed due to %s:", err.Error())
		}
		config.MessageQueue.Port = 9889
		inData, _ := yaml.Marshal(&config)
		cfg.WriteViperConfig(inData)
	}
}
