package main

import (
	"log"
	"path/filepath"
	"selfHttp"
	"strconv"
	"bytes"
	"mime/multipart"
	"io"
	"os"
	"util"
	"strings"
	"net/http"
	"compress/gzip"
	"config"
)

type uploadRequest struct {
	file    string
	url     string
	count   int
}

type uploader struct {
	cfg         *config.CliJsonConfig
	logNotify   chan uploadRequest
	reportor    *FailureReporter
}

func NewUploader(cfg *config.CliJsonConfig) *uploader {
	uploader := uploader{}
	uploader.cfg       = cfg
	uploader.logNotify = make(chan uploadRequest)
	uploader.reportor  = NewFailureReporter(cfg)

	return &uploader
}

func (l *uploader) LogNotify(file, url string, count int) {
	var req uploadRequest
	req.file  = file
	req.url   = url
	req.count = count

	l.logNotify <- req
}

func (l *uploader) UploadLog() {
	for {
		select {
		case req := <- l.logNotify:
			{
				dir := filepath.Dir(req.file)
				basename := filepath.Base(req.file)
				log.Printf("upload log, dir:%v, name:%v", dir, basename)

				go l.uploadlogFile(dir, basename, req.url, req.count)
			}
		}
		//case runningLogfile := <- l.runningLogNotify:
		//	{
		//		log.Printf("upload running log file: %s", runningLogfile)
		//
		//		basename := filepath.Base(runningLogfile)
		//		go l.uploadlogFile(cfg.GetRunningLogPath(), basename, cfg.GetUploadRunningLogServerUrl())
		//	}
		//case crashLogfile := <- l.crashLogNotify:
		//	{
		//		log.Printf("upload crash log file: %s", crashLogfile)
		//
		//		basename := filepath.Base(crashLogfile)
		//		go l.uploadlogFile(cfg.GetCrashLogPath(), basename, cfg.GetUploadCrashLogServerUrl())
		//	}
		//}
	}
}

func (l *uploader) setUploadBasicParm(w *multipart.Writer, timestamp, logname string) bool {
	var err error

	err = w.WriteField("apptype", l.cfg.GetAppType())
	if err != nil {
		log.Printf("Fail write app_type field, error:%s", err.Error())
		return false
	}

	err = w.WriteField("appkey", l.cfg.GetAppKey())
	if err != nil {
		log.Printf("Fail write app_key field, error:%s", err.Error())
		return false
	}

	err = w.WriteField("appid", l.cfg.GetAppId())
	if err != nil {
		log.Printf("Fail write app_id field, error:%s", err.Error())
		return false
	}

	err = w.WriteField("timestamp", timestamp)
	if err != nil {
		log.Printf("Fail write timestamp field, error:%s", err.Error())
		return false
	}

	if util.IsRunningLog(logname) {
		logname = strings.TrimSuffix(logname, ".log")
	}
	err = w.WriteField("logname", logname)
	if err != nil {
		log.Printf("Fail write logname field, error:%s", err.Error())
		return false
	}

	if _, ok := APP_TYPE_MAP[l.cfg.GetAppType()]; !ok {
		log.Printf("Unknown app_type")
		return false
	}
	signature := util.GenerateSignature(APP_TYPE_MAP[l.cfg.GetAppType()], logname, SCERET, timestamp)
	err = w.WriteField("signature", signature)
	if err != nil {
		log.Printf("Fail write signature field, error:%s", err.Error())
		return false
	}

	return true
}

//func (l *uploader) queryLogExist(client *http.Client, deviceid, logname string) (bool, error) {
//	if util.IsRunningLog(logname) {
//		logname = strings.TrimSuffix(logname, ".log")
//	}
//
//	params := make(map[string]string)
//	params["deviceid"] = deviceid
//	params["logname"] = logname
//
//	resp, err := selfHttp.HttpGet(client, l.cfg.GetQueryLogExistUrl(), params)
//	if err != nil {
//		return false, err
//	}
//
//	if string(resp) == "EXIST" {
//		return true, nil
//	}
//
//	return false, nil
//}

func (l *uploader) getSrvTimestamp(client *http.Client) (string, error) {
	logTimestampUrl := l.cfg.GetLogTimestampUrl()

	resp, err := selfHttp.HttpGet(client, logTimestampUrl, map[string]string{})
	if err != nil {
		return "", err
	}

	_, err = strconv.ParseInt(string(resp), 10, 64)
	if err != nil {
		return "", err
	}

	return string(resp), nil
}

func (l *uploader) compressFile(filename string) (*bytes.Buffer, error) {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("compress file, open fail, path:%s, error:%s", filename, err.Error())
		return nil, err
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	zw := gzip.NewWriter(buf)

	filestat, err := file.Stat()
	if err != nil {
		log.Printf("compress file, stat fail, path:%s, error:%s", filename, err.Error())
		return nil, err
	}

	zw.Name = filestat.Name()
	zw.ModTime = filestat.ModTime()
	written, err := io.Copy(zw, file)
	if err != nil {
		log.Printf("compress file, copy fail, path:%s, error:%s", filename, err.Error())
		return nil, err
	}
	log.Printf("compress file, path:%s, written:%d", filename, written)

	err = zw.Flush()
	if err != nil {
		log.Printf("compress file, flush fail, path:%s, error:%s", filename, err.Error())
		return nil, err
	}
	err = zw.Close()
	if err != nil {
		log.Printf("compress file, close fail, path:%s, error:%s", filename, err.Error())
		return nil, err
	}

	return buf, nil
}

func (l *uploader) uploadlogFile(path, logname, url string, count int) {
	absFileName := path + string(os.PathSeparator) + logname

	// 创建http客户端
	httpClient := selfHttp.HttpRequestClient(l.cfg.GetHttpConnnectTimeout(),
		                                     l.cfg.GetHttpRWTimeout(),
		                                     l.cfg.GetHttpsCaCertPath())

	// 查询服务端日志是否存在
	//exist, err := l.queryLogExist(httpClient, l.cfg.GetAppId(), logname)
	//if err != nil {
	//	log.Printf("Http query log exist fail, error:%s", err.Error())
	//	return
	//}
	//
	//if exist {
	//	log.Printf("Log file already exist, log:%s", logname)
	//	err := os.Remove(absFileName) // TODO 直接删除
	//	if err != nil {
	//		log.Printf("Remove file fail, error:%s", err.Error())
	//	}
	//	return
	//}

	// 请求服务器时间戳
	timestamp, err := l.getSrvTimestamp(httpClient)
	if err != nil {
		log.Printf("Http get srv timestamp fail, error:%s", err.Error())
		return
	}

	// 压缩数据
	gzipbuf, err := l.compressFile(absFileName)
	if err != nil {
		log.Printf("Compress file fail")
		return
	}

	// 封装请求
	//fh, err := os.Open(absFileName)
	//if err != nil {
	//	log.Printf("Open file fail, error:%s", err.Error())
	//	return
	//}
	//defer fh.Close()

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	if !l.setUploadBasicParm(w, timestamp, logname) {
		log.Printf("Set upload post param fail")
		return
	}

	fw, err := w.CreateFormFile("uploadfile", logname)
	if err != nil {
		log.Printf("Fail create form file, error:%s", err.Error())
		return
	}

	//written, err := io.Copy(fw, fh)
	written, err := io.Copy(fw, gzipbuf)
	if err != nil {
		log.Printf("Fail Copy, error:%s", err.Error())
		return
	}
	w.Close()

	log.Printf("Success copy, file:%s, written:%d", absFileName, written)

	appid := l.cfg.GetAppId()

	// 执行
	_, err = selfHttp.HttpPost(httpClient, url, buf, w.FormDataContentType())
	if err != nil {
		log.Printf("Fail http post, error:%s", err.Error())
		l.reportor.ReportUploadResult(l.cfg.GetLogFailureRateUrl(), appid, logname, count, "fail")
		return
	}

	l.reportor.ReportUploadResult(l.cfg.GetLogFailureRateUrl(), appid, logname, count, "succ")
	// 删除文件
	os.Remove(absFileName)

	return
}
