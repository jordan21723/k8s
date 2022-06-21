package downloader

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	config "k8s-installer/pkg/config/downloader"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/pkg/util/netutils"
)

type Downloader interface {
	Run(ctx context.Context) error
	// RunStream(ctx context.Context) (io.Reader, error)
	Cleanup()
}

func CalculateTimeout(cfg *config.DownloadConfig) time.Duration {
	if cfg == nil {
		return constants.DefaultDownloadTimeout
	}
	// the timeout specified by user should be used firstly
	if cfg.Timeout > 0 {
		return cfg.Timeout
	}
	timeout := netutils.CalculateTimeout(cfg.RV.FileLength, cfg.MinRate,
		constants.DefaultMinRate, 10*time.Second)
	if timeout > 0 {
		return timeout
	}
	return constants.DefaultDownloadTimeout
}

func DoDownloadTimeout(downloader Downloader, timeout time.Duration) error {
	if timeout <= 0 {
		logrus.Warnf("invaild download timeout(%.3fs), use default:(%.3f)",
			timeout.Seconds(), constants.DefaultDownloadTimeout.Seconds())
		timeout = constants.DefaultDownloadTimeout
	}
	ctx, cancel := context.WithCancel(context.Background())

	var ch = make(chan error)
	go func() {
		ch <- downloader.Run(ctx)
	}()
	defer cancel()

	var err error
	select {
	case err = <-ch:
		return err
	case <-time.After(timeout):
		err = fmt.Errorf("download timeout (%.3fs)", timeout.Seconds())
		downloader.Cleanup()
	}
	return err
}

func MoveFile(src string, dst string, expectMd5 string) error {
	start := time.Now()
	if expectMd5 != "" {
		realMd5 := fileutils.Md5Sum(src)
		logrus.Infof("compute raw md5:%s for file:%s cost:%.3fs", realMd5,
			src, time.Since(start).Seconds())
		if realMd5 != expectMd5 {
			return fmt.Errorf("md5NotMacth, real:%s expect %s", realMd5, expectMd5)
		}
	}
	err := fileutils.MoveFile(src, dst)
	logrus.Infof("move src:%s to dst:%s result:%t cost:%.3f",
		src, dst, err == nil, time.Since(start).Seconds())
	return err
}
