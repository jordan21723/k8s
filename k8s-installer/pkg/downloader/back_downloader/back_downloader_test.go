package downloader

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"gopkg.in/check.v1"
	config "k8s-installer/pkg/config/downloader"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/util/fileutils"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type BackDownloaderTestSuite struct {
	workHome string
	host     string
	ln       net.Listener
}

func init() {
	check.Suite(&BackDownloaderTestSuite{})
}

func (s *BackDownloaderTestSuite) SetUpSuite(c *check.C) {
	s.workHome, _ = ioutil.TempDir("/tmp", "caas-BackDownloaderTestSuite-")
	s.host = fmt.Sprintf("127.0.0.1:%d", rand.Intn(1000)+63000)
	s.ln, _ = net.Listen("tcp", s.host)
	go http.Serve(s.ln, http.FileServer(http.Dir(s.workHome)))
}

func (s *BackDownloaderTestSuite) TearDownSuite(c *check.C) {
	if s.workHome != "" {
		if err := os.RemoveAll(s.workHome); err != nil {
			fmt.Printf("remove path:%s error", s.workHome)
		}
	}
	s.ln.Close()
}

func (s *BackDownloaderTestSuite) TestBackDownloader_Run(c *check.C) {
	testFileMd5 := fileutils.CreateTestFileWithMD5(filepath.Join(s.workHome, "download.test"), "test downloader")
	dst := filepath.Join(s.workHome, "back.test")

	cfg := CreateConfig(nil, s.workHome)
	bd := &BackDownloader{
		cfg:    cfg,
		URL:    "http://" + s.host + "/download.test",
		Target: dst,
	}

	bd.cleaned = false
	cfg.BackSourceReason = constants.BackSourceReasonNoSpace
	c.Assert(bd.Run(context.TODO()), check.NotNil)

	cfg.BackSourceReason = 0
	bd.cleaned = false
	c.Assert(bd.Run(context.TODO()), check.IsNil)

	// test: realMd5 doesn't equal to expectedMd5
	bd.cleaned = false
	bd.Md5 = testFileMd5
	c.Assert(bd.Run(context.TODO()), check.IsNil)
	md5sum := fileutils.Md5Sum(dst)
	c.Assert(md5sum, check.Equals, bd.Md5)

	// test: realMd5 equals to expectedMd5
	bd.cleaned = false
	bd.Md5 = testFileMd5
	c.Assert(bd.Run(context.TODO()), check.IsNil)

	// test: target not exists
	bd.Md5 = testFileMd5
	bd.Target = filepath.Join(s.workHome, "back1.test")
	c.Assert(bd.Run(context.TODO()), check.IsNil)
	md5sum = fileutils.Md5Sum(bd.Target)
	c.Assert(md5sum, check.Equals, bd.Md5)
}

func (s *BackDownloaderTestSuite) TestBackDownloader_Run_NotExist(c *check.C) {
	dst := filepath.Join(s.workHome, "back.test")

	cfg := CreateConfig(nil, s.workHome)
	bd := &BackDownloader{
		cfg:    cfg,
		URL:    "http://" + s.host + "/download1.test",
		Target: dst,
	}

	err := bd.Run(context.TODO())
	c.Check(err, check.NotNil)
	c.Check(err, check.ErrorMatches, ".*404")
}

func CreateConfig(writer io.Writer, workHome string) *config.DownloadConfig {
	if writer == nil {
		writer = ioutil.Discard
	}
	cfg := config.NewDownloadConfig()

	logrus.StandardLogger().Out = writer
	return cfg
}
