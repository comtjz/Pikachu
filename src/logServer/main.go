package main

import (
	"tools"
	"os"
	"log"
	"net/http"
	"selfHttp"
	"time"
)

func main() {
	// 启动删除日志作业
	delLogTask := cfgJson.GetDelLogTask()
	for _, task := range delLogTask {
		remover := tools.NewFileCycleRemover(task, &cfgJson)
		go remover.Start()
	}

	http.HandleFunc("/", sayHello)
	http.HandleFunc("/xpanec/timestamp", getSrvTimeStamp)
	http.HandleFunc("/xpanec/log-failure-rate", recordUploadResult)

	// 设置接收日志接口
	saveLogTask := cfgJson.GetSaveLogTask()
	for _, task := range saveLogTask {
		ok, method := cfgJson.GetSaveLogMethod(task)
		if !ok {
			continue
		}
		ok, logDir := cfgJson.GetSaveLogLogPath(task)
		if !ok {
			continue
		}
		log.Printf("method:%s, path:%s", method, logDir)
		handler := NewReceiveLogHandler(task, logDir, &cfgJson)
		http.Handle(method, handler)
	}

	//http.HandleFunc("/", sayHello)
	//http.HandleFunc("/xpanec/timestamp", getSrvTimeStamp)
	//http.HandleFunc("/xpanec/log/upload", receiveLogFile)
	//http.HandleFunc("/xpanec/crash/upload", receiveLogFile)
	//http.HandleFunc("/xpanec/log-exist", isExistLogFile)

	pid := os.Getpid()
	address := cfgJson.GetHttpSrvAddr()

	log.Printf("process with pid %d serving %s.\n", pid, address)
	readTime  := time.Duration(cfgJson.GetHttpReadTimeout()) * time.Second
	writeTime := time.Duration(cfgJson.GetHttpWriteTimeout()) * time.Second
	//err := selfHttp.NewServer(address, nil,
	//	time.Duration(cfgJson.GetHttpReadTimeout()) * time.Second, time.Duration(cfgJson.GetHttpWriteTimeout()) * time.Second).ListenAndServe()

	certFile := cfgJson.GetHttpsCertFile()
	keyFile := cfgJson.GetHttpsKeyFile()
	err := selfHttp.NewServer(address, nil, readTime, writeTime).ListenAndServeTLS(certFile, keyFile)
	log.Printf("process with pid %d stoped, error: %s.\n", pid, err.Error())
}