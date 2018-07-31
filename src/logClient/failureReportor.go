package main

import (
	"selfHttp"
	"time"
	"net/url"
	"strings"
	"log"
	"strconv"
	"config"
)

type FailureReporter struct {
	cfg  *config.CliJsonConfig
}

func NewFailureReporter(cfg *config.CliJsonConfig) *FailureReporter {
	reportor := &FailureReporter{}
	reportor.cfg = cfg

	return reportor
}

func (f *FailureReporter)ReportUploadResult(failureUrl, appid, filename string, count int, result string) error {

	sCount := strconv.FormatInt(int64(count), 10)
	// 创建http客户端
	httpClient := selfHttp.HttpRequestClient(10, 20, f.cfg.GetHttpsCaCertPath())

	var failureInfo = url.Values{}
	failureInfo.Add("appid", appid)
	failureInfo.Add("filename", filename)
	failureInfo.Add("count", sCount)
	failureInfo.Add("result", result)

	data := failureInfo.Encode()
	_, err := selfHttp.HttpPost(httpClient, failureUrl, strings.NewReader(data), "application/x-www-form-urlencoded")
	if err != nil {
		time.Sleep(time.Second * 1)
		_, err = selfHttp.HttpPost(httpClient, failureUrl, strings.NewReader(data), "application/x-www-form-urlencoded")
		if err != nil {
			log.Printf("report upload result fail, error:%s", err.Error())
			return err
		}
	}

	return nil
}
