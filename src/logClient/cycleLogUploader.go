package main

import (
	"os"
	"util"
	"time"
	"io/ioutil"
	"log"
	"strings"
	"config"
)

type uploadRecord struct {
	uploadCount    int64
	lastUploadTime int64
}

type cycleLogUploader struct {
	task         string
	cfg          *config.CliJsonConfig
	uploader     *uploader
	records      map[string]uploadRecord
}

func NewCycleLogUploader(task string, cfg *config.CliJsonConfig, uploader *uploader) *cycleLogUploader {
	cycleLogUploader := &cycleLogUploader{}
	cycleLogUploader.task     = task
	cycleLogUploader.cfg      = cfg
	cycleLogUploader.uploader = uploader
	cycleLogUploader.records  = make(map[string]uploadRecord)

	return cycleLogUploader
}

func (c *cycleLogUploader)Start() {
	log.Printf("cycle log uploader start, taskname:%s", c.task)
	for {
		ok, logDir := c.cfg.GetUploadLogLogDir(c.task)
		if !ok {
			log.Printf("cycle log uploader game over, taskname:%s", c.task)
			return
		}

		ok, logServerUrl := c.cfg.GetUploadLogLogServerUrl(c.task)
		if !ok {
			log.Printf("cycle log uploader game over, taskname:%s", c.task)
			return
		}

		ok, interval := c.cfg.GetUploadLogInterval(c.task)
		if !ok {
			log.Printf("cycle log uploader game over, taskname:%s", c.task)
			return
		}
		if interval < 30 * 60 { // TODO 循环间隔最短半小时
			interval = 30 * 60
		}

		log.Printf("cycle log uploader process, task:%s", c.task)

		startTime := time.Now().Unix()

		noUploadLogFiles := c.searchNoUploadLogFile(logDir)
		if len(noUploadLogFiles) > 0 {
			for _, noUploadLogFile := range noUploadLogFiles {
				record, _ := c.records[noUploadLogFile]
				record.uploadCount += 1
				record.lastUploadTime = time.Now().Unix()

				c.uploader.LogNotify(noUploadLogFile, logServerUrl, int(record.uploadCount))
			}
		}

		c.amendUploadRecord(logDir)

		endTime := time.Now().Unix()
		if (endTime - startTime < int64(interval)) {
			time.Sleep(time.Second * time.Duration(int64(interval) - (endTime - startTime)))
		}
	}
}

//func NewCycleLogUploader(uploader *uploader, isRunningLog bool) *cycleLogUploader {
//	cycleLogUploader := cycleLogUploader{}
//	cycleLogUploader.uploader     = uploader
//	cycleLogUploader.isRunningLog = isRunningLog
//	cycleLogUploader.records      = make(map[string]uploadRecord)
//
//	return &cycleLogUploader
//}

// 注意：不进入子目录
func (ru *cycleLogUploader) searchNoUploadLogFile(path string) []string {
	noUploadLogFiles := make([]string, 0)

	curTime := time.Now().Unix()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("Read Dir Error: path:%s, error:%s", path, err.Error())
		return noUploadLogFiles
	}

	for _, fi := range files {
		if fi.IsDir() {
			continue
		}

		if strings.HasPrefix(fi.Name(), ".") {
			continue
		}

		log.Printf("Path:%s, file:%s", path, fi.Name())
		if (curTime - fi.ModTime().Unix() > 24 * 60 * 60) { // 一天
			continue
		}

		if (curTime - fi.ModTime().Unix() < 30 * 60) { // 半小时
			continue
		}

		if !util.IsRunningLog(fi.Name()) && !util.IsCrashLog(fi.Name()) {
			continue
		}

		filename := fi.Name()
		if record, ok := ru.records[filename]; ok {
			if record.uploadCount > 5 {
				continue
			}
		} else {
			ru.records[filename] = uploadRecord{}
		}

		noUploadLogFiles = append(noUploadLogFiles, filename)
	}

	return noUploadLogFiles
}

func (ru *cycleLogUploader) amendUploadRecord(path string) {
	for filename, _ := range ru.records {
		absfilename := path + string(os.PathSeparator) + filename
		if exist, _ := util.PathExists(absfilename); !exist {
			log.Printf("Delete upload record, path:%s", absfilename)
			delete(ru.records, filename)
		}
	}
}

//func (ru *cycleLogUploader)CycleUploadLogFile() {
//	for {
//		if ru.isRunningLog {
//			log.Printf("Start cycle running log upload")
//		} else {
//			log.Printf("Start cycle crash log upload")
//		}
//
//		startTime := time.Now().Unix()
//
//		noUploadLogFiles := ru.searchNoUploadLogFile()
//		if len(noUploadLogFiles) > 0 {
//			for _, noUploadLogFile := range noUploadLogFiles {
//				record, _ := ru.records[noUploadLogFile]
//				record.uploadCount += 1
//				record.lastUploadTime = time.Now().Unix()
//
//				if ru.isRunningLog {
//					ru.uploader.RunningLogNotify(noUploadLogFile)
//				} else {
//					ru.uploader.CrashLogNotify(noUploadLogFile)
//				}
//			}
//		}
//
//		ru.amendUploadRecord()
//
//		endTime := time.Now().Unix()
//		cycleUploadLogInterval := cfg.GetCycleUploadLogInterval()
//		if (endTime - startTime < cycleUploadLogInterval) {
//			time.Sleep(time.Second * time.Duration(cycleUploadLogInterval - (endTime - startTime)))
//		}
//	}
//}