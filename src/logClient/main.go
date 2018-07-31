package main

import (
	"os"
	"os/signal"
	"syscall"
	"log"
	"tools"
	"util"
	"crypto/md5"
	"fmt"
	"config"
)

func setupSignal() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGUSR1)

	go func() {
		for sig := range c {
			log.Printf("receive signal: %v", sig)
			if sig == syscall.SIGUSR1 {
				var tmpCfg config.CliJsonConfig
				tmpCfg.LoadJsonConfig(cfgJsonFile)

				if _, isExist := APP_TYPE_MAP[tmpCfg.GetAppType()]; !isExist {
					log.Println("Wrong Config, Unknow App Type")
					continue
				}

				mac_addr := util.MacAddr()
				if len(mac_addr) == 0 {
					log.Println("Fail Get Mac Addr")
					continue
				}

				md5_mac := md5.Sum([]byte(mac_addr[0]))
				md5_mac_str := fmt.Sprintf("%x", md5_mac)
				device_id := util.GenerateDeviceId(APP_TYPE_MAP[tmpCfg.GetAppType()], md5_mac_str)

				tmpCfg.SetAppKey(md5_mac_str)
				tmpCfg.SetAppId(device_id)

				if err := cfgJson.CopyConfig(&tmpCfg); err != nil {
					log.Println("Fail Copy Config")
				}
				continue
			}

			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				os.Exit(0)
			}
		}
	}()
}

func main() {
	setupSignal()

	// 启动删除日志作业
	delLogTask := cfgJson.GetDelLogTask()
	for _, task := range delLogTask {
		remover := tools.NewFileCycleRemover(task, &cfgJson)
		go remover.Start()

		logRemover[task] = remover
	}

	//启动日志上传器
	logUploader := NewUploader(&cfgJson)
	go func() {
		logUploader.UploadLog()
		uploadHandlerDone <- true
	}()

	// TEST START
	//testUploader := NewTestUploader(logUploader)
	//testUploader.testGetSrvTimestamp()
	//testUploader.testIsExistLogFile()
	//testUploader.testRuningLogUploader("这就是搜索引擎：核心技术详解.pdf")
	//testUploader.testCrashLogUploader("搬运工厂www.bygc945.xin -专注免费精品教程分享.url")
	// TEST END

	// 启动循环检查触发的日志上报作业
	roundRobinTask := cfgJson.GetRounRobinUploadLogTask()
	for _, task := range roundRobinTask {
		cycleLogUploader := NewCycleLogUploader(task, &cfgJson, logUploader)
		go cycleLogUploader.Start()
	}

	// 启动文件创建触发的日志上报作业
	fileNotifyTask := cfgJson.GetFileNotifyUploadLogTask()
	for _, task := range fileNotifyTask {
		fnLogUploader := NewFileNotifyLogUploader(task, &cfgJson, logUploader)
		go fnLogUploader.Start()
	}

	<-uploadHandlerDone
}
