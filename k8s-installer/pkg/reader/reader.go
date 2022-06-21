package reader

import (
	"crypto/md5"
	"hash"
	"io"

	"k8s-installer/pkg/util/fileutils"
)

func NewFileReader(src io.Reader, calculateMd5 bool) *FileReader {
	var md5sum hash.Hash
	if calculateMd5 {
		md5sum = md5.New()
	}

	return &FileReader{
		Src:    src,
		md5sum: md5sum,
	}
}

type FileReader struct {
	Src    io.Reader
	md5sum hash.Hash
}

func (r *FileReader) Read(p []byte) (n int, err error) {
	n, e := r.Src.Read(p)
	if e != nil && e != io.EOF {
		return n, e
	}
	if n > 0 {
		if r.md5sum != nil {
			r.md5sum.Write(p[:n])
		}
	}
	return n, e
}

func (r *FileReader) Md5() string {
	if r.md5sum != nil {
		return fileutils.GetMd5Sum(r.md5sum, nil)
	}
	return ""
}
