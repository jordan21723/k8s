package fileutils

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/dep"
	"k8s-installer/pkg/log"
	"k8s-installer/pkg/server/version"
	"k8s-installer/pkg/util/stringutils"
	"k8s-installer/schema"

	"github.com/hyperboloide/lk"
	"github.com/shirou/gopsutil/disk"
	"gopkg.in/yaml.v2"

	funk "github.com/thoas/go-funk"
)

// BufferSize defines the buffer size when reading and writing file.
const BufferSize = 8 * 1024 * 1024

var _localMacAddressList = getMacAddressList()

func CreateTestFileWithMD5(path string, content string) string {
	f, err := createFile(path, content)
	if err != nil {
		log.Errorf("fail to create file %v due to error: %v", f, err)
		return ""
	}
	defer f.Close()
	return Md5Sum(f.Name())
}

func createFile(path string, content string) (*os.File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if content != "" {
		f.WriteString(content)
	}
	return f, nil
}

// CreateDirectory creates directory recursively.
func CreateDirectory(dirPath string) error {
	f, e := os.Stat(dirPath)
	if e != nil {
		if os.IsNotExist(e) {
			return os.MkdirAll(dirPath, 0755)
		}
		return fmt.Errorf("failed to create dir %s: %v", dirPath, e)
	}
	if !f.IsDir() {
		return fmt.Errorf("failed to create dir %s: dir path already exists and is not a directory", dirPath)
	}
	return e
}

// DeleteFile deletes a file not a directory.
func DeleteFile(filePath string) error {
	if !PathExist(filePath) {
		return fmt.Errorf("failed to delete file %s: file not exist", filePath)
	}
	if IsDir(filePath) {
		return fmt.Errorf("failed to delete file %s: file path is a directory rather than a file", filePath)
	}
	return os.Remove(filePath)
}

// DeleteFiles deletes all the given files.
func DeleteFiles(filePaths ...string) {
	if len(filePaths) > 0 {
		for _, f := range filePaths {
			DeleteFile(f)
		}
	}
}

func DeleteDirectory(dir string) error {
	if !PathExist(dir) {
		return nil
	}
	return os.RemoveAll(dir)
}

// OpenFile opens a file. If the parent directory of the file isn't exist,
// it will create the directory.
func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	if PathExist(path) {
		return os.OpenFile(path, flag, perm)
	}
	if err := CreateDirectory(filepath.Dir(path)); err != nil {
		return nil, err
	}

	return os.OpenFile(path, flag, perm)
}

// Link creates a hard link pointing to src named linkName for a file.
func Link(src string, linkName string) error {
	if PathExist(linkName) {
		if IsDir(linkName) {
			return fmt.Errorf("failed to link %s to %s: link name already exists and is a directory", linkName, src)
		}
		if err := DeleteFile(linkName); err != nil {
			return fmt.Errorf("failed to link %s to %s when deleting target file: %v", linkName, src, err)
		}

	}
	return os.Link(src, linkName)
}

// SymbolicLink creates target as a symbolic link to src.
func SymbolicLink(src string, target string) error {
	if !PathExist(src) {
		return fmt.Errorf("failed to symlink %s to %s: src no such file or directory", target, src)
	}
	if PathExist(target) {
		if IsDir(target) {
			return fmt.Errorf("failed to symlink %s to %s: link name already exists and is a directory", target, src)
		}
		if err := DeleteFile(target); err != nil {
			return fmt.Errorf("failed to symlink %s to %s when deleting target file: %v", target, src, err)
		}

	}
	return os.Symlink(src, target)
}

// CopyFile copies the file src to dst.
func CopyFile(src string, dst string) (err error) {
	var (
		s *os.File
		d *os.File
	)
	if !IsRegularFile(src) {
		return fmt.Errorf("failed to copy %s to %s: src is not a regular file", src, dst)
	}
	if s, err = os.Open(src); err != nil {
		return fmt.Errorf("failed to copy %s to %s when opening source file: %v", src, dst, err)
	}
	defer s.Close()

	if PathExist(dst) && !IsDir(dst) {
		if err := DeleteFile(dst); err != nil {
			return fmt.Errorf("failed to copy %s to %s when copy dst file: %v", src, dst, err)
		}
	}
	if d, err = OpenFile(dst, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0755); err != nil {
		return fmt.Errorf("failed to copy %s to %s when opening destination file: %v", src, dst, err)
	}
	defer d.Close()

	buf := make([]byte, BufferSize)
	for {
		n, err := s.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to copy %s to %s when reading src file: %v", src, dst, err)
		}
		if n == 0 || err == io.EOF {
			break
		}
		if _, err := d.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to copy %s to %s when writing dst file: %v", src, dst, err)
		}
	}
	return nil
}

// MoveFile moves the file src to dst.
func MoveFile(src string, dst string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("failed to move %s to %s: src is not a regular file", src, dst)
	}
	if PathExist(dst) && !IsDir(dst) {
		if err := DeleteFile(dst); err != nil {
			return fmt.Errorf("failed to move %s to %s when deleting dst file: %v", src, dst, err)
		}
	}
	return os.Rename(src, dst)
}

// MoveFileAfterCheckMd5 will check whether the file's md5 is equals to the param md5
// before move the file src to dst.
func MoveFileAfterCheckMd5(src string, dst string, md5 string) error {
	if !IsRegularFile(src) {
		return fmt.Errorf("failed to move file with md5 check %s to %s: src is not a regular file", src, dst)
	}
	m := Md5Sum(src)
	if m != md5 {
		return fmt.Errorf("failed to move file with md5 check %s to %s: md5 of source file doesn't match against the given md5 value", src, dst)
	}
	return MoveFile(src, dst)
}

// PathExist reports whether the path is exist.
// Any error get from os.Stat, it will return false.
func PathExist(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// IsDir reports whether the path is a directory.
func IsDir(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.IsDir()
}

// IsRegularFile reports whether the file is a regular file.
// If the given file is a symbol link, it will follow the link.
func IsRegularFile(name string) bool {
	f, e := os.Stat(name)
	if e != nil {
		return false
	}
	return f.Mode().IsRegular()
}

// Md5Sum generates md5 for a given file.
func Md5Sum(name string) string {
	if !IsRegularFile(name) {
		return ""
	}
	f, err := os.Open(name)
	if err != nil {
		return ""
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, BufferSize)
	h := md5.New()

	_, err = io.Copy(h, r)
	if err != nil {
		return ""
	}

	return GetMd5Sum(h, nil)
}

// GetMd5Sum gets md5 sum as a string and appends the current hash to b.
func GetMd5Sum(md5 hash.Hash, b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}

// GetSys returns the underlying data source of the os.FileInfo.
func GetSys(info os.FileInfo) (*syscall.Stat_t, bool) {
	sys, ok := info.Sys().(*syscall.Stat_t)
	return sys, ok
}

// LoadYaml loads yaml config file.
func LoadYaml(path string, out interface{}) error {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load yaml %s when reading file: %v", path, err)
	}
	if err = yaml.Unmarshal(content, out); err != nil {
		return fmt.Errorf("failed to load yaml %s: %v", path, err)
	}
	return nil
}

// GetFreeSpace gets the free disk space of the path.
func GetFreeSpace() (constants.Size, error) {
	fs := syscall.Statfs_t{}
	var totalSize constants.Size
	parts, _ := disk.Partitions(true)
	for _, p := range parts {
		device := p.Mountpoint

		if err := syscall.Statfs(device, &fs); err != nil {
			return 0, err
		}
		totalSize += constants.Size(fs.Bavail * uint64(fs.Bsize))
	}

	return totalSize, nil
}

// IsEmptyDir check whether the directory is empty.
func IsEmptyDir(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if _, err = f.Readdirnames(1); err == io.EOF {
		return true, nil
	}
	return false, err
}

func WalkDir(root string) ([]string, error) {
	var output []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !IsDir(path) {
			output = append(output, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

func WalkMatch(root, pattern string) ([]string, error) {
	var matches []string
	f, err := os.Open(root)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	for _, f := range list {
		if matched, err := filepath.Match(pattern, filepath.Base(f.Name())); err != nil {
			return nil, err
		} else if matched {
			matches = append(matches, f.Name())
		}
	}

	return matches, nil
}

func AppendFile(filename string, text string, block string) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	count := 0
	var newContent []string
	for scanner.Scan() {
		s := scanner.Text()
		if strings.Contains(s, block) {
			count = count + 1
			continue
		}
		if (count % 2) != 0 {
			continue
		}
		newContent = append(newContent, s)
	}

	newContent = append(newContent, text)

	t := strings.Join(newContent, "\n")

	if _, err = f.WriteString(t); err != nil {
		return err
	}
	return nil
}

func FileExists(path string) error {
	_, err := os.Stat(path)
	return err
}

func Untar(file, dir string) error {
	r, err := os.Open(file)
	if err != nil {
		return err
	}
	defer r.Close()
	t0 := time.Now()
	nFiles := 0
	madeDir := map[string]bool{}
	defer func() {
		td := time.Since(t0)
		if err == nil {
			log.Debugf("extracted tarball into %s: %d files, %d dirs (%v)", dir, nFiles, len(madeDir), td)
		} else {
			log.Debugf("error extracting tarball into %s after %d files, %d dirs, %v: %v", dir, nFiles, len(madeDir), td, err)
		}
	}()
	zr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("requires gzip-compressed body: %v", err)
	}
	tr := tar.NewReader(zr)
	loggedChtimesError := false
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Debugf("tar reading error: %v", err)
			return fmt.Errorf("tar error: %v", err)
		}
		if !validRelPath(f.Name) {
			return fmt.Errorf("tar contained invalid name error %q", f.Name)
		}
		rel := filepath.FromSlash(f.Name)
		abs := filepath.Join(dir, rel)

		fi := f.FileInfo()
		mode := fi.Mode()
		switch {
		case mode.IsRegular():
			// Make the directory. This is redundant because it should
			// already be made by a directory entry in the tar
			// beforehand. Thus, don't check for errors; the next
			// write will fail with the same error.
			dir := filepath.Dir(abs)
			if !madeDir[dir] {
				if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
					return err
				}
				madeDir[dir] = true
			}
			wf, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}
			n, err := io.Copy(wf, tr)
			if closeErr := wf.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
			if err != nil {
				return fmt.Errorf("error writing to %s: %v", abs, err)
			}
			if n != f.Size {
				return fmt.Errorf("only wrote %d bytes to %s; expected %d", n, abs, f.Size)
			}
			modTime := f.ModTime
			if modTime.After(t0) {
				// Clamp modtimes at system time. See
				// golang.org/issue/19062 when clock on
				// buildlet was behind the gitmirror server
				// doing the git-archive.
				modTime = t0
			}
			if !modTime.IsZero() {
				if err := os.Chtimes(abs, modTime, modTime); err != nil && !loggedChtimesError {
					// benign error. Gerrit doesn't even set the
					// modtime in these, and we don't end up relying
					// on it anywhere (the gomote push command relies
					// on digests only), so this is a little pointless
					// for now.
					log.Debugf("error changing modtime: %v (further Chtimes errors suppressed)", err)
					loggedChtimesError = true // once is enough
				}
			}
			nFiles++
		case mode.IsDir():
			if err := os.MkdirAll(abs, 0755); err != nil {
				return err
			}
			madeDir[abs] = true
		default:
			return fmt.Errorf("tar file entry %s contained unsupported file type %v", f.Name, mode)
		}
	}
	return nil
}

func validRelativeDir(dir string) bool {
	if strings.Contains(dir, `\`) || path.IsAbs(dir) {
		return false
	}
	dir = path.Clean(dir)
	if strings.HasPrefix(dir, "../") || strings.HasSuffix(dir, "/..") || dir == ".." {
		return false
	}
	return true
}

func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}

func WriteToFile(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("file create failed. err: " + err.Error())
	} else {
		// offset
		//os.Truncate(filename, 0) //clear
		n, _ := f.Seek(0, os.SEEK_END)
		_, err = f.WriteAt([]byte(content), n)
		fmt.Println("write succeed!")
		defer f.Close()
	}
	return err
}

func GetDepList(dep dep.DepMap) []string {
	var res []string
	for _, v := range dep["centos"][constants.V1_18_6]["x86_64"] {
		res = append(res, strings.Replace(v, ".rpm", "", 1))
	}
	return res
}

func ParseLicense(licenseStr string) (*schema.SystemInfo, error) {
	result := schema.SystemInfo{
		SystemProduct: constants.SystemInfoProduct,
		SystemVersion: constants.SystemInfoVersion + "." + version.Version,
	}

	publicKey, err := lk.PublicKeyFromB32String(constants.Publickey)
	if err != nil {
		result.ErrorDetail = fmt.Sprintf("illegal license %v", err)
		return &result, err
	}

	lic, err := lk.LicenseFromB32String(licenseStr)
	if err != nil {
		result.ErrorDetail = fmt.Sprintf("illegal license %v", err)
		return &result, err
	}

	if ok, err := lic.Verify(publicKey); err != nil {
		result.ErrorDetail = fmt.Sprintf("illegal license %v", err)
		return &result, err
	} else if !ok {
		result.ErrorDetail = fmt.Sprintf("illegal license %v", err)
		return &result, nil
	}

	if err := json.Unmarshal(lic.Data, &result); err != nil {
		return &result, err
	}

	if err := validateLicense(&result); err != nil {
		return &result, err
	}

	return &result, nil
}

func validateLicense(lic *schema.SystemInfo) error {
	sysVer := stringutils.SplitTrimFilter(constants.SystemInfoVersion, ".")
	licVer := stringutils.SplitTrimFilter(lic.Version, ".")
	if constants.SystemInfoProduct != lic.Product || len(sysVer) < 1 || len(licVer) < 1 || sysVer[0] != licVer[0] {
		lic.ErrorDetail = "license product or version invalid"
		return nil
	}
	if lic.Expired.Before(time.Now()) {
		lic.ErrorDetail = "license expired"
		return nil
	}
	node, cpu, err := getNodeCab()
	if err != nil {
		return err
	}
	if node > stringutils.S(lic.Node).DefaultInt(node) {
		lic.ErrorDetail = "node number exceed limit"
		return nil
	}
	if cpu > stringutils.S(lic.CPU).DefaultInt(cpu) {
		lic.ErrorDetail = "cpu number exceed limit"
		return nil
	}

	licMacAddressList := stringutils.SplitTrimFilter(lic.MacAddress, " ")
	licMacAddressList = funk.Map(licMacAddressList, func(s string) string {
		return strings.ToLower(s)
	}).([]string)
	if len(licMacAddressList) > 0 {
		result := funk.Join(_localMacAddressList, licMacAddressList, funk.InnerJoin).([]string)
		if len(result) < 1 {
			lic.ErrorDetail = fmt.Sprintf("Local mac address (%#v) does not match the license (%#v)", _localMacAddressList, licMacAddressList)
			return nil
		}
	}

	lic.LicenseValid = true
	return nil
}

func getMacAddressList() (macAddressList []string) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("fail to get net interfaces: %v", err)
		return macAddressList
	}

	macAddressList = funk.Map(netInterfaces, func(n net.Interface) string {
		return strings.ToLower(n.HardwareAddr.String())
	}).([]string)
	macAddressList = funk.Filter(macAddressList, func(s string) bool {
		return len(s) > 0
	}).([]string)

	return macAddressList
}

func getNodeCab() (int, int, error) {

	runtimeCache := cache.GetCurrentCache()
	nodeInfoCollection, errGetNode := runtimeCache.GetNodeInformationCollection()
	if errGetNode != nil {
		return 0, 0, errGetNode
	}

	var cpuCoreCount uint
	for _, node := range nodeInfoCollection {
		cpuCoreCount += node.SystemInfo.CPU.Cores
	}

	return len(nodeInfoCollection), int(cpuCoreCount), nil
}
