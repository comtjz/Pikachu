package tools

import (
	"log"
	"time"
	"io/ioutil"
	"os"
	"config"
)

type FileCycleRemover struct {
	task  string
	cfg   config.DelLog
}

func NewFileCycleRemover(task string, cfg config.DelLog) *FileCycleRemover {
	remover := &FileCycleRemover{}
	remover.task = task
	remover.cfg  = cfg

	return remover
}

func (r *FileCycleRemover)Start() {
	log.Printf("file cycle remover start, taskname:%s", r.task)
	for {
		ok, path := r.cfg.GetDelLogLogDir(r.task)
		if !ok {
			log.Printf("file cycle remover game over, taskname:%s", r.task)
			return
		}

		ok, expiredTime := r.cfg.GetDelLogLogExpiredTime(r.task)
		if !ok {
			log.Printf("file cycle remover game over, taskname:%s", r.task)
			return
		}

		if expiredTime < 24 * 60 * 60 { // TODO 必须保留一天
			expiredTime = 24 * 60 * 60
		}

		ok, cycleTime := r.cfg.GetDelLogInterval(r.task)
		if !ok {
			log.Printf("file cycle remover game over, taskname:%s", r.task)
			return
		}

		if cycleTime <= 0 { // TODO 间隔为负值或0
			cycleTime = 60 * 5
		}

		log.Printf("cycle remove expired log file, task:%s, path:%s, expiredTime:%d, cycleTime:%d",
			r.task, path, expiredTime, cycleTime)

		startTime := time.Now().Unix()

		expiredLogFiles := searchExpiredLogFile(path, int64(expiredTime))
		if len(expiredLogFiles) > 0 {
			delLogFileOrDir(expiredLogFiles, path)
		}

		emptyLogDir := searchEmptyLogDir(path)
		if len(emptyLogDir) > 0 {
			delLogFileOrDir(emptyLogDir, path)
		}

		endTime := time.Now().Unix()
		if (endTime - startTime < int64(cycleTime)) {
			time.Sleep(time.Second * time.Duration(int64(cycleTime) - (endTime - startTime)))
		}
	}
}

func delLogFileOrDir(delFilesOrDirs []string, outPath string) {
	for _, delFileOrDir := range delFilesOrDirs {
		if delFileOrDir == outPath {
			continue
		}

		err := os.Remove(delFileOrDir)
		if err != nil {
			log.Printf("File Remove Error, filename:%s, error:%s", delFileOrDir, err.Error())
		} else {
			log.Printf("File Remove OK, filename:%s", delFileOrDir)
		}
	}
}

func searchEmptyLogDir(path string) []string {
	emptyLogDir := make([]string, 0)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("Read Dir Error, path:%s, error:%s", path, err.Error())
		return emptyLogDir
	}

	if len(files) == 0 {
		emptyLogDir = append(emptyLogDir, path)
		return emptyLogDir
	}

	for _, fi := range files {
		subPathName := path + string(os.PathSeparator) + fi.Name()
		if fi.IsDir() {
			subEmptyLogDir := searchEmptyLogDir(subPathName)
			if len(subEmptyLogDir) > 0 {
				emptyLogDir = append(emptyLogDir, subEmptyLogDir...)
			}
		}
	}

	return emptyLogDir
}

func searchExpiredLogFile(path string, expiredTime int64) []string {
	expiredLogFiles := make([]string, 0)

	curTime := time.Now().Unix()
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Printf("Read Dir Error, path:%s, error:%s", path, err.Error())
		return expiredLogFiles
	}

	for _, fi := range files {
		subPathName := path + string(os.PathSeparator) + fi.Name()
		if fi.IsDir() {
			subExpiredLogFiles := searchExpiredLogFile(subPathName, expiredTime)
			if len(subExpiredLogFiles) > 0 {
				expiredLogFiles = append(expiredLogFiles, subExpiredLogFiles...)
			}
		} else {
			modTime := fi.ModTime().Unix()
			if (curTime - modTime > expiredTime) {
				expiredLogFiles = append(expiredLogFiles, subPathName)
			}
		}
	}

	return expiredLogFiles
}
