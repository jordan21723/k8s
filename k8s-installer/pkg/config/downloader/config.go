package config

import (
	"fmt"
	"os"
	"time"

	"k8s-installer/pkg/constants"
)

type RuntimeVariable struct {
	RealTarget string
	FileLength int64
}

type DownloadConfig struct {
	URL string

	Timeout time.Duration `json:"timeout,omitempty"`
	Md5     string

	StartTime time.Time `json:"-"`
	// sign =  'Pid + float64(time.Now().UnixNano))/float64(time.Second) format: "%d-%.3f"'.
	Sign string `json:"-"`

	RV               RuntimeVariable `json:"-"`
	BackSourceReason int             `json:"-"`

	LocalLimit constants.Size `yaml:"localLimit,omitempty" json:"localLimit,omitempty"`
	MinRate    constants.Size `yaml:"minRate,omitempty" json:"minRate,omitempty"`
}

func NewDownloadConfig() *DownloadConfig {
	cfg := new(DownloadConfig)
	cfg.StartTime = time.Now()
	cfg.Sign = fmt.Sprintf("%d-%.3f",
		os.Getpid(), float64(time.Now().UnixNano())/float64(time.Second))

	cfg.RV.FileLength = -1
	return cfg
}
