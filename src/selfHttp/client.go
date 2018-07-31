package selfHttp

import (
	"net/http"
	"net"
	"time"
	"io"
	"log"
	"io/ioutil"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"util"
)

func HttpRequestClient(connTimeout ,rwTimeout int, caCertPath string) *http.Client {
	tr := &http.Transport{
		MaxIdleConnsPerHost: 10,
		Dial: func(netw, addr string) (net.Conn, error) {
			conn, err := net.DialTimeout(netw, addr, time.Duration(connTimeout) * time.Second)
			if err != nil {
				return nil, err
			}

			conn.SetDeadline(time.Now().Add(time.Duration(rwTimeout) * time.Second))

			return conn, nil
		},
		DisableKeepAlives: true,
	}

	if util.IsFile(caCertPath) {
		pool := x509.NewCertPool()
		caCrt, _ := ioutil.ReadFile(caCertPath)
		pool.AppendCertsFromPEM(caCrt)
		tr.TLSClientConfig = &tls.Config{RootCAs: pool}
	}

	return &http.Client{
		Transport: tr,
	}
}

func HttpGet(client *http.Client, url string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("Fail http request, error:%s", err.Error())
		return nil, err
	}

	values := req.URL.Query()
	for k, v := range params {
		values.Add(k, v)
	}
	req.URL.RawQuery = values.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Fail http request, error:%s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	//if resp.StatusCode != http.StatusOK {
	//	log.Printf("Invalid HTTP Code，%d", resp.StatusCode)
	//	return nil, fmt.Errorf("Invalid HTTP Code: [%v]", resp.StatusCode)
	//}

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Fail read response, error:%s", err.Error())
		return nil, err
	}

	return resp_body, nil
}

func HttpPost(client *http.Client, url string, buf io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		log.Println("Fail new request, error:%s", err.Error())
		return nil, err
	}

	// TODO 头部信息
	req.Header.Set("Content-Type", contentType)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Fail http request, error:%s", err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Fail read response, error:%s", err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Invalid HTTP Code，%v, response:%v", resp.StatusCode, string(resp_body))
		return nil, fmt.Errorf("Invalid HTTP Code: [%v]", resp.StatusCode)
	}

	return resp_body, nil
}
