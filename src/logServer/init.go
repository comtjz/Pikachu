package main

import (
	"flag"
	"fmt"
	"config"
	"tools"
)

var (
	//cfg       config.SrvConfig
	//cfgFile   string

	cfgJson     config.SrvJsonConfig
	cfgJsonFile string

	APP_TYPE_MAP map[string]string

	rateLimiter *tools.RateLimiter
	statistican *UploadStatistician
)

const (
	PREFIX_LOG_2_IP = "xpanec_l2ip"
	PREFIX_ID_2_LOG = "xpanec_id2l"

	SCERET = "d565ab9469347c97968849968913fb65"  // md5 -s PH_F4-xpanec
)

func init() {
	flag.StringVar(&cfgJsonFile, "c", "config/srv_conf.json", "please set config file!")
	flag.Parse()

	var err error
	if err = cfgJson.LoadJsonConfig(cfgJsonFile); err != nil {
		panic( fmt.Sprintf("init: parsing cnnfig error '%s'", err.Error()) )
	}

	APP_TYPE_MAP = make(map[string]string)
	APP_TYPE_MAP["wkyun"] = "201807121531385579"

	rateLimiter = tools.NewRateLimiter()
	statistican = NewUploadStatistician()
}
