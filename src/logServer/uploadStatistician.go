package main

import (
	"sync"
	"time"
	"strconv"
)

type statisticalInfo struct {
	firstSuccess     int64
	firstFail        int64

	retrySuccess     int64
}

type UploadStatistician struct {
	mtx sync.Mutex

	uploadFile map[string]map[string]bool  // 记录当前小时内收到的上报文件
	statistics map[string]*statisticalInfo  // 记录当前小时内上报文件成功失败
}

func NewUploadStatistician() *UploadStatistician {
	uploadStatistician := &UploadStatistician{}
	uploadStatistician.uploadFile = make(map[string]map[string]bool)
	uploadStatistician.statistics = make(map[string]*statisticalInfo)

	go uploadStatistician.cleanupStatisticalInfo()

	return uploadStatistician
}

func (u *UploadStatistician) RecordUploadResult(id_file, count ,result string) {
	isFirstUpload := false
	iCount, _ := strconv.ParseInt(count, 10, 64)
	if iCount == 1 {
		isFirstUpload = true
	}

	isSuccess := false
	if result == "succ" {
		isSuccess = true
	}

	u.mtx.Lock()
	defer u.mtx.Unlock()

	curTm := time.Now().Format("2006-01-02 15")

	if _, ok := u.uploadFile[curTm]; !ok {
		u.uploadFile[curTm] = make(map[string]bool)
	}

	if _, ok := u.uploadFile[curTm][id_file]; !ok {
		u.uploadFile[curTm][id_file] = true
	}

	if _, ok := u.statistics[curTm]; !ok {
		u.statistics[curTm] = &statisticalInfo{}
	}

	if isFirstUpload {
		if isSuccess {
			u.statistics[curTm].firstSuccess += 1
		} else {
			u.statistics[curTm].firstFail += 1
		}
	} else {
		if isSuccess {
			u.statistics[curTm].retrySuccess += 1
		}
	}

	return
}

func (u *UploadStatistician) cleanupStatisticalInfo() {
	for {
		time.Sleep(time.Minute * 10)
		u.mtx.Lock()

		for key, _ := range u.uploadFile {
			tm, err := time.Parse("2006-01-02 15", key)
			if err != nil {
				delete(u.uploadFile, key)
			}

			if time.Now().Sub(tm) > time.Hour * 24 {
				delete(u.uploadFile, key)
			}
		}

		for key, _ := range u.statistics {
			tm, err := time.Parse("2006-01-02 15", key)
			if err != nil {
				delete(u.statistics, key)
			}

			if time.Now().Sub(tm) > time.Hour * 24 {
				delete(u.uploadFile, key)
			}
		}
		u.mtx.Unlock()
	}
}