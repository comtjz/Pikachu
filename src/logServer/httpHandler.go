package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"os"
	"fmt"
	"time"
	"strconv"
	"util"
	"config"
)

func renderError(response http.ResponseWriter, message string, statusCode int) {
	response.WriteHeader(statusCode)
	response.Write([]byte(message))
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		renderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Hello World!")
	fmt.Fprintf(w, "Hello World!")
}

func getSrvTimeStamp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		renderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	curTime := fmt.Sprintf("%d", time.Now().Unix())
	w.Write([]byte(curTime))

	return
}

func recordUploadResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		renderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	appid := r.PostFormValue("appid")
	filename := r.PostFormValue("filename")
	count := r.PostFormValue("count")
	result := r.PostFormValue("result")

	id_file := appid + "_" + filename
	statistican.RecordUploadResult(id_file, count, result)

	return
}

func isExistLogFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		renderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()

	deviceid := r.FormValue("deviceid")
	logname  := r.FormValue("logname")
	log.Printf("Query log exist, deviceid:%s, logname:%s", deviceid, logname)
	if deviceid == "" || logname == "" {
		renderError(w, "INVALID_PARAMETER", http.StatusBadRequest)
		return
	}

	//if !util.IsRunningLog(logname) {
	//	renderError(w, "INVALID_PARAMETER", http.StatusBadRequest)
	//	return
	//}

	// 通过redis查询日志文件是否存在
	exist := false
	if !exist {
		w.Write([]byte("NO_EXIST"))
		return
	}

	w.Write([]byte("EXIST"))

	return
}

func authUploader(r *http.Request) bool {
	apptype := r.PostFormValue("apptype")
	if apptype == "" {
		log.Println("Empty app_type")
		return false
	}

	appkey := r.PostFormValue("appkey")
	if appkey == "" {
		log.Println("Empty app_key")
		return false
	}

	appid := r.PostFormValue("appid")
	if appid == "" {
		log.Println("Empty app_id")
		return false
	}

	timestamp := r.PostFormValue("timestamp")
	if timestamp == "" {
		log.Println("Empty timestamp")
		return false
	}

	logname := r.PostFormValue("logname")
	if logname == "" {
		log.Println("Empty logname")
		return false
	}

	signature := r.PostFormValue("signature")
	if signature == "" {
		log.Println("Empty signature")
		return false
	}

	// 验证时间戳
	tm, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		log.Println("Unknown timestamp")
		return false
	}

	curTm := time.Now().Unix()
	if curTm < tm || curTm - tm > 180 { // 三分钟内的请求可以处理
		log.Printf("Wrong timestamp, curTm:%d, tm:%d", curTm, tm)
		return false
	}

	// 验证device id
	if _, ok := APP_TYPE_MAP[apptype]; !ok {
		log.Println("Unknown app_type")
		return false
	}

	real_app_type := APP_TYPE_MAP[apptype]
	log.Printf("app_type:%s, appkey:%s", real_app_type, appkey)
	deviceId := util.GenerateDeviceId(real_app_type, appkey)
	if deviceId == "" || deviceId != appid {
		log.Printf("Wrong app_id, deviceId:%s, appId:%s", deviceId, appid)
		return false
	}

	// 验证auth
	deviceSign := util.GenerateSignature(real_app_type, logname, SCERET, timestamp)
	if deviceSign == "" || deviceSign != signature {
		log.Println("Wrong signature")
		return false
	}

	return true
}

type receiveLogHandler struct {
	task     string
	logDir   string
	cfg      *config.SrvJsonConfig
}

func NewReceiveLogHandler(task, logDir string, cfg *config.SrvJsonConfig) *receiveLogHandler {
	handler := &receiveLogHandler{}
	handler.task   = task
	handler.logDir = logDir
	handler.cfg    = cfg

	return handler
}

func (h *receiveLogHandler)ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("method : %v", r.Method)
	if r.Method != "POST" {
		renderError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(32<<20) // 32MB

	if !authUploader(r) {
		renderError(w, "INVALID_PARAMETER", http.StatusBadRequest)
		return
	}

	appId := r.PostFormValue("appid")
	logName := r.PostFormValue("logname")
	log.Printf("Receive log file, appid:%s, logname:%s", appId, logName)

	//if util.IsRunningLog(logName) {
	//	logName += ".log"
	//}

	/* 控制访问频率 */
	limiter := rateLimiter.GetVisitor(appId)
	if limiter.Allow() == false {
		renderError(w, http.StatusText(429), http.StatusTooManyRequests)
		return
	}

	/* 处理上传日志 */
	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		log.Println("Invalid File")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println("Invalid Read File")
		renderError(w, "INVALID_FILE", http.StatusBadRequest)
	}

	// write file
	path := h.logDir + string(os.PathSeparator) + appId + string(os.PathSeparator)
	os.MkdirAll(path, 0777)
	path += logName

	newFile, err := os.Create(path)
	if err != nil {
		log.Println("Invalid Create File, error:" + err.Error())
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		log.Println("Can't Write File")
		renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
		return
	}

	// 记录上传

	return
}
