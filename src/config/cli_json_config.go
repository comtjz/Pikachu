package config

import (
	"sync"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"util"
	"errors"
)

type CliDelLogUnit struct {
	LogDir          string
	LogExpiredTime  int
	Interval        int
}

type UploadLogUnit struct {
	LogDir          string
	LogServerUrl    string
	CheckType       string
	Interval        int
}

type CliJsonConfig struct {
	AppType              string
	HttpConnectTimeout   int
	HttpRWTimeout        int
	HttpsCaCertPath      string
	LogTimestampUrl      string
	LogFailureRateUrl    string
	DelLog               map[string]CliDelLogUnit
	UploadLog            map[string]UploadLogUnit

	mutex    sync.Mutex
	appKey   string
	deviceId string
}

func (c *CliJsonConfig) LoadJsonConfig(cfgFile string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	data, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *CliJsonConfig) GetAppType() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.AppType
}

func (c *CliJsonConfig) GetHttpsCaCertPath() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpsCaCertPath
}

func (c *CliJsonConfig) GetHttpConnnectTimeout() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpConnectTimeout
}

func (c *CliJsonConfig) GetHttpRWTimeout() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpRWTimeout
}

func (c *CliJsonConfig) GetLogTimestampUrl() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.LogTimestampUrl
}

func (c *CliJsonConfig) GetLogFailureRateUrl() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.LogFailureRateUrl
}

func (c *CliJsonConfig) GetDelLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delLog := make([]string, 0)
	for key, _ := range c.DelLog {
		delLog = append(delLog, key)
	}

	return delLog
}

func (c *CliJsonConfig) GetDelLogLogDir(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.LogDir
	}

	return false, ""
}

func (c *CliJsonConfig) GetDelLogLogExpiredTime(task string) (bool, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.LogExpiredTime
	}

	return false, 0
}

func (c *CliJsonConfig) GetDelLogInterval(task string) (bool, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.Interval
	}

	return false, 0
}

func (c *CliJsonConfig) GetUploadLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	uploadLog := make([]string, 0)
	for key, _ := range c.UploadLog {
		uploadLog = append(uploadLog, key)
	}

	return uploadLog
}

func (c *CliJsonConfig) GetFileNotifyUploadLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	fileNotify := make([]string, 0)
	for key, val := range c.UploadLog {
		if val.CheckType == "file-notify" {
			fileNotify = append(fileNotify, key)
		}
	}

	return fileNotify
}

func (c *CliJsonConfig) GetRounRobinUploadLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	roundRobin := make([]string, 0)
	for key, val := range c.UploadLog {
		if val.CheckType == "round-robin" {
			roundRobin = append(roundRobin, key)
		}
	}

	return roundRobin
}

func (c *CliJsonConfig) GetUploadLogLogDir(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.UploadLog[task]; ok {
		return true, job.LogDir
	}

	return false, ""
}

func (c *CliJsonConfig) GetUploadLogLogServerUrl(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.UploadLog[task]; ok {
		return true, job.LogServerUrl
	}

	return false, ""
}

func (c *CliJsonConfig) GetUploadLogCheckType(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.UploadLog[task]; ok {
		return true, job.CheckType
	}

	return false, ""
}

func (c *CliJsonConfig) GetUploadLogInterval(task string) (bool, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.UploadLog[task]; ok {
		return true, job.Interval
	}

	return false, 0
}

func (c *CliJsonConfig) SetAppKey(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.appKey = key
}

func (c *CliJsonConfig) GetAppKey() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.appKey
}

func (c *CliJsonConfig) SetAppId(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.deviceId = id
}

func (c *CliJsonConfig) GetAppId() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.deviceId
}

func (c *CliJsonConfig) CopyConfig(tmp *CliJsonConfig) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	tmp.mutex.Lock()
	defer tmp.mutex.Unlock()

	fnUploadLogTask := make([]string, 0)
	for key, val := range tmp.UploadLog {
		if val.CheckType == "file-notify" {
			fnUploadLogTask = append(fnUploadLogTask, key)
		}
	}

	rrUploadLogTask := make([]string, 0)
	for key, val := range tmp.UploadLog {
		if val.CheckType == "round-robin" {
			rrUploadLogTask = append(rrUploadLogTask, key)
		}
	}

	if len(fnUploadLogTask) == 0 && len(rrUploadLogTask) == 0 {
		return errors.New("init: no log upload job is set")
	}

	for key, val := range tmp.UploadLog {
		if val.LogDir == "" {
			return errors.New( fmt.Sprintf("init: log dir is empty or not exist, task:%s", key) )
		}

		if exist, _ := util.PathExists(val.LogDir); !exist {
			return errors.New( fmt.Sprintf("init: log dir is not exist, task:%s", key) )
		}
	}

	c.AppType            = tmp.AppType
	c.HttpConnectTimeout = tmp.HttpConnectTimeout
	c.HttpRWTimeout      = tmp.HttpRWTimeout
	c.UploadLog = map[string]UploadLogUnit{}
	for key, val := range tmp.UploadLog {
		c.UploadLog[key] = val
	}
	c.DelLog = map[string]CliDelLogUnit{}
	for key, val := range tmp.DelLog {
		c.DelLog[key] = val
	}
	c.appKey   = tmp.appKey
	c.deviceId = tmp.deviceId

	return nil
}