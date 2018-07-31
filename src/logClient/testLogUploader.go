package main

import (
	"selfHttp"
	"log"
)

type testUploader struct {
	uploader *uploader
}

func NewTestUploader(uploader *uploader) *testUploader {
	tu := testUploader{}
	tu.uploader = uploader

	return &tu
}

func (tu *testUploader)testGetSrvTimestamp() {
	// 创建http客户端
	httpClient := selfHttp.HttpRequestClient(cfgJson.GetHttpConnnectTimeout(),
		                                     cfgJson.GetHttpRWTimeout(),
		                                     cfgJson.GetHttpsCaCertPath())

	timestamp, err := tu.uploader.getSrvTimestamp(httpClient)
	if err != nil {
		log.Printf("http get srv timestamp fail, error:%v", err.Error())
		return
	}

	log.Printf("server timestamp: %s", timestamp)

	return
}

//func (tu *testUploader)testIsExistLogFile() {
//	// 创建http客户端
//	httpClient := selfHttp.HttpRequestClient(time.Duration(cfgJson.GetHttpConnnectTimeout()),
//		                                     time.Duration(cfgJson.GetHttpRWTimeout()))
//
//	exist, err := tu.uploader.queryLogExist(httpClient, cfgJson.GetAppId(), "BaiduKernel_20180718141312_671_1.log")
//	if err != nil {
//		log.Printf("http query log exist fail, error:%v", err.Error())
//		return
//	}
//
//	log.Printf("log exist: %v", exist)
//
//	return
//}

func (tu *testUploader)testRuningLogUploader(file string, url string) {
	tu.uploader.LogNotify(file, url, 1)
}
