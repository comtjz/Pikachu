package main

import (
	"flag"
	"fmt"
	"util"
	"config"
	"crypto/md5"
	"log"
	"tools"
)

const (
	SCERET = "d565ab9469347c97968849968913fb65"  // md5 -s PH_F4-xpanec
)

var (
	cfgJson   config.CliJsonConfig
	cfgJsonFile string

	logRemover map[string]*tools.FileCycleRemover

	// watchLog chan
	watcherHandlerDone = make(chan bool)
	// fileHandleMainJob chan
	fileHandlerDone = make(chan bool)
	// uploadhandle chan
	uploadHandlerDone = make(chan bool)

	APP_TYPE_MAP map[string]string
)

func init() {
	flag.StringVar(&cfgJsonFile, "c", "config/cli_conf.json", "please set config json file!")
	flag.Parse()

	var err error
	if err = cfgJson.LoadJsonConfig(cfgJsonFile); err != nil {
		panic( fmt.Sprintf("init: parsing configJson error '%s'", err.Error()))
	}

	fnUploadLogTask := cfgJson.GetFileNotifyUploadLogTask()
	rrUploadLogTask := cfgJson.GetRounRobinUploadLogTask()
	if len(fnUploadLogTask) == 0 && len(rrUploadLogTask) == 0 {
		panic( fmt.Sprintf("init: no log upload job is set") )
	}

	for _, task := range fnUploadLogTask {
		_, logDir := cfgJson.GetUploadLogLogDir(task)
		if logDir == "" {
			panic( fmt.Sprintf("init: log dir is empty or not exist, task:%s", task) )
		}

		if exist, _ := util.PathExists(logDir); !exist {
			panic( fmt.Sprintf("init: log dir is not exist, task:%s", task) )
		}
	}

	for _, task := range rrUploadLogTask {
		_, logDir := cfgJson.GetUploadLogLogDir(task)
		if logDir == "" {
			panic( fmt.Sprintf("init: log dir is empty or not exist, task:%s", task) )
		}

		if exist, _ := util.PathExists(logDir); !exist {
			panic( fmt.Sprintf("init: log dir is not exist, task:%s", task) )
		}
	}

	logRemover = make(map[string]*tools.FileCycleRemover)

	APP_TYPE_MAP = make(map[string]string)
	APP_TYPE_MAP["wkyun"] = "201807121531385579"

	if _, isExist := APP_TYPE_MAP[cfgJson.GetAppType()]; !isExist {
		panic("wrong app type")
	}

	mac_addr := util.MacAddr()
	if len(mac_addr) == 0 {
		panic("empty mac address")
	}

	md5_mac := md5.Sum([]byte(mac_addr[0]))
	md5_mac_str := fmt.Sprintf("%x", md5_mac)
	log.Printf("apptype:%s, md5_mac_str:%s", APP_TYPE_MAP[cfgJson.GetAppType()], md5_mac_str)
	device_id := util.GenerateDeviceId(APP_TYPE_MAP[cfgJson.GetAppType()], md5_mac_str)

	cfgJson.SetAppKey(md5_mac_str)
	cfgJson.SetAppId(device_id)
}