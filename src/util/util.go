package util

import (
	"os"
	"strings"
	"net"
	"io/ioutil"
	"sort"
	"fmt"
	"crypto/md5"
	"path/filepath"
	"encoding/hex"
)

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func IsDir(dirname string) bool {
	fhandler, err := os.Stat(dirname)
	if (!(err == nil || os.IsExist(err))) {
		return false;
	} else {
		return fhandler.IsDir()
	}
}

func IsFile(filename string) bool {
	fhandler, err := os.Stat(filename)
	if (!(err == nil || os.IsExist(err))) {
		return false
	} else if (fhandler.IsDir()) {
		return false
	}

	return true
}

func GetFileByteSize(filename string) (bool, int64) {
	if (!IsFile(filename)) {
		return false, 0
	}

	fhandler, err := os.Stat(filename)
	if (err != nil) {
		return false, 0
	}
	return true, fhandler.Size()
}

// 获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配前缀，后缀过滤
// 返回结果按mtime排序
func ListDir(dirPath, prefix, suffix string) (files []os.FileInfo, err error) {
	files = make([]os.FileInfo, 0, 10)

	dir, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(dir, func(i, j int) bool { return dir[i].ModTime().UnixNano() < dir[j].ModTime().UnixNano() })

	prefix = strings.ToLower(prefix)
	suffix = strings.ToLower(suffix)

	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}

		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		fileName := strings.ToLower(fi.Name())
		if prefix != "" && !strings.HasPrefix(fileName, prefix) {
			continue
		}

		if suffix != "" && !strings.HasSuffix(fileName, suffix) {
			continue
		}

		files = append(files, fi)
	}

	return files, nil
}

func MacAddr() []string {
	macInterfaces := make([]string, 0)
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	for _, inter := range interfaces {
		mac := inter.HardwareAddr
		macInterfaces = append(macInterfaces, string(mac))
	}

	return macInterfaces
}

func GenerateDeviceId(deviceType string, deviceAddr string) string {
	source := fmt.Sprintf("%s#%s", deviceType, deviceAddr)

	src := md5.Sum([]byte(source))

	high1 := uint64(src[3])<<24 | uint64(src[2])<<16 | uint64(src[1])<<8 | uint64(src[0])
	high2 := uint64(src[7])<<24 | uint64(src[6])<<16 | uint64(src[5])<<8 | uint64(src[4])
	high3 := uint64(src[11])<<24 | uint64(src[10])<<16 | uint64(src[9])<<8 | uint64(src[8])
	high4 := uint64(src[15])<<24 | uint64(src[14])<<16 | uint64(src[13])<<8 | uint64(src[12])
	sign1 := high1 + high3
	sign2 := high2 + high4
	sign := sign1&0xFFFFFFFF | sign2<<32

	deviceId := fmt.Sprintf("%018d", sign)
	deviceId = deviceId[:18]
	return deviceId
}

func GenerateSignature(deviceType, logname, secret, timestamp string) string {
	source := fmt.Sprintf("%s#%s#%s#%s", deviceType, logname, secret, timestamp)

	ctx := md5.New()
	ctx.Write([]byte(source))

	return hex.EncodeToString(ctx.Sum(nil))
}

func IsRunningLog(path string) bool {
	basename := filepath.Base(path)
	if strings.HasPrefix(basename, "BaiduKernel_") {
		return true
	}
	return false
}

func IsCrashLog(path string) bool {
	//basename := filepath.Base(path)
	// TODO

	return true
}