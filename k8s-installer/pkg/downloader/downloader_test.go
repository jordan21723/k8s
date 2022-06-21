package downloader

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/check.v1"
	"k8s-installer/pkg/util/fileutils"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type DownloaderTestSuite struct {
}

func init() {
	check.Suite(&DownloaderTestSuite{})
}

func (s *DownloaderTestSuite) TestDoDownloadTimeout(c *check.C) {
	md := &MockDownloader{100}

	err := DoDownloadTimeout(md, 0*time.Millisecond)
	c.Assert(err, check.IsNil)

	err = DoDownloadTimeout(md, 50*time.Millisecond)
	c.Assert(err, check.NotNil)

	err = DoDownloadTimeout(md, 110*time.Millisecond)
	c.Assert(err, check.IsNil)
}

func (s *DownloaderTestSuite) TestMoveFile(c *check.C) {
	tmp, _ := ioutil.TempDir("/tmp", "caas-test")
	defer os.RemoveAll(tmp)

	src := filepath.Join(tmp, "a")
	dst := filepath.Join(tmp, "b")
	md5str := fileutils.CreateTestFileWithMD5(src, "hello")

	err := MoveFile(src, dst, "x")
	c.Assert(fileutils.PathExist(src), check.Equals, true)
	c.Assert(fileutils.PathExist(dst), check.Equals, false)
	c.Assert(err, check.NotNil)

	err = MoveFile(src, dst, md5str)
	c.Assert(fileutils.PathExist(src), check.Equals, false)
	c.Assert(fileutils.PathExist(dst), check.Equals, true)
	c.Assert(err, check.IsNil)
	content, _ := ioutil.ReadFile(dst)
	c.Assert(string(content), check.Equals, "hello")

	err = MoveFile(src, dst, "")
	c.Assert(err, check.NotNil)
}

type MockDownloader struct {
	Sleep int
}

func (md *MockDownloader) Run(ctx context.Context) error {
	time.Sleep(time.Duration(md.Sleep) * time.Millisecond)
	return nil
}

func (md *MockDownloader) Cleanup() {
}
