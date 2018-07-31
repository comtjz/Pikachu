package config

import (
	"sync"
	"io/ioutil"
	"encoding/json"
)

type SaveLogUnit struct {
	Method   string
	LogPath  string
}

type SrvDelLogUnit struct {
	LogDir          string
	LogExpiredTime  int
	Interval        int
}

type SrvJsonConfig struct {
	HttpSrvAddr      string
	HttpReadTimeout  int
	HttpWriteTimeout int
	HttpsCertFile    string
	HttpsKeyFile     string
	SaveLog          map[string]SaveLogUnit
	DelLog           map[string]SrvDelLogUnit

	mutex  sync.Mutex
}

func (c *SrvJsonConfig) LoadJsonConfig(cfgFile string) error {
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

func (c *SrvJsonConfig) GetHttpSrvAddr() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpSrvAddr
}

func (c *SrvJsonConfig) GetHttpReadTimeout() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpReadTimeout
}

func (c *SrvJsonConfig) GetHttpWriteTimeout() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpWriteTimeout
}

func (c *SrvJsonConfig) GetHttpsCertFile() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpsCertFile
}

func (c *SrvJsonConfig) GetHttpsKeyFile() string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.HttpsKeyFile
}

func (c *SrvJsonConfig) GetSaveLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	saveLogTask := make([]string, 0)
	for key, _ := range c.SaveLog {
		saveLogTask = append(saveLogTask, key)
	}

	return saveLogTask
}

func (c *SrvJsonConfig) GetSaveLogMethod(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.SaveLog[task]; ok {
		return true, job.Method
	}

	return false, ""
}

func (c *SrvJsonConfig) GetSaveLogLogPath(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.SaveLog[task]; ok {
		return true, job.LogPath
	}

	return false, ""
}

func (c *SrvJsonConfig) GetDelLogTask() []string {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delLog := make([]string, 0)
	for key, _ := range c.DelLog {
		delLog = append(delLog, key)
	}

	return delLog
}

func (c *SrvJsonConfig) GetDelLogLogDir(task string) (bool, string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.LogDir
	}

	return false, ""
}

func (c *SrvJsonConfig) GetDelLogLogExpiredTime(task string) (bool, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.LogExpiredTime
	}

	return false, 0
}

func (c *SrvJsonConfig) GetDelLogInterval(task string) (bool, int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if job, ok := c.DelLog[task]; ok {
		return true, job.Interval
	}

	return false, 0
}
