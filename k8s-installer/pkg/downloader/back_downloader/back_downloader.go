package downloader

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	config "k8s-installer/pkg/config/downloader"
	"k8s-installer/pkg/config/printer"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/downloader"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/reader"
	"k8s-installer/pkg/util"
	"k8s-installer/pkg/util/fileutils"
	"k8s-installer/pkg/util/stringutils"
)

type BackDownloader struct {
	URL          string
	Target       string
	Md5          string
	TaskID       string
	cfg          *config.DownloadConfig
	tempFileName string
	cleaned      bool
}

var _ downloader.Downloader = &BackDownloader{}

func NewBackDownloader(cfg *config.DownloadConfig) *BackDownloader {
	var (
		taskID string
	)
	return &BackDownloader{
		cfg:    cfg,
		URL:    cfg.URL,
		Target: cfg.RV.RealTarget,
		Md5:    cfg.Md5,
		TaskID: taskID,
	}
}

func (bd *BackDownloader) isSuccessStatus(code int) bool {
	return code < 400
}

func (bd *BackDownloader) Run(ctx context.Context) error {
	var (
		resp *http.Response
		err  error
		f    *os.File
	)
	if bd.cfg.BackSourceReason == constants.BackSourceReasonNoSpace {
		bd.cfg.BackSourceReason += constants.ForceNotBackSourceAddition
		err = fmt.Errorf("download fail and not back source: %d", bd.cfg.BackSourceReason)
		return err
	}

	defer bd.Cleanup()

	prefix := "backsource." + bd.cfg.Sign + "."
	if f, err = ioutil.TempFile(filepath.Dir(bd.Target), prefix); err != nil {
		return err
	}
	bd.tempFileName = f.Name()
	defer f.Close()

	if resp, err = util.HttpGet(bd.URL, 0); err != nil {
		return err
	}
	defer resp.Body.Close()

	if !bd.isSuccessStatus(resp.StatusCode) {
		return fmt.Errorf("failed to download from source, response code:%d", resp.StatusCode)
	}

	buf := make([]byte, 512*1024)
	reader := reader.NewFileReader(resp.Body, bd.Md5 != "")
	if _, err = io.CopyBuffer(f, reader, buf); err != nil {
		return err
	}

	var realMd5 string
	if _, err := os.Stat(bd.Target); err != nil {
		if os.IsNotExist(err) {
			realMd5 = ""
		}
	} else {
		realMd5 = fileutils.Md5Sum(bd.Target)
	}
	if bd.Md5 == "" || bd.Md5 != realMd5 {
		printer.Printf("start downdload %s from the source station", filepath.Base(bd.Target))
		log.Debugf("Try to download file from %s(%s) to %s(%s)", bd.URL, bd.Md5, bd.Target, realMd5)
		err = downloader.MoveFile(bd.tempFileName, bd.Target, "")
	} else {
		log.Debugf("%s md5 match %s, skip download.", bd.URL, bd.Target)
	}
	return err
}

// Cleanup all temporary resources
func (bd *BackDownloader) Cleanup() {
	if bd.cleaned {
		return
	}
	if !stringutils.IsEmptyStr(bd.tempFileName) {
		fileutils.DeleteFile(bd.tempFileName)
	}
	bd.cleaned = true
}
