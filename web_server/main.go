package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"time"
)

var configFilePath string

type Config struct {
	DbHost           string `json:"dbhost"`
	DbPort           int    `json:"dbport"`
	DbUser           string `json:"dbuser"`
	DbPwd            string `json:"dbpwd"`
	ServerListenPort int    `json:"listenport"`
}

func main() {
	flag.StringVar(&configFilePath, "c", "config.json", "config file path")
	flag.Parse()
	ts := time.Now().Format("20060102150405")
	initLog(ts + "server.log")
	if b, err := ioutil.ReadFile(configFilePath); err != nil {
		LogError("read config file error %s", err.Error())
		return
	} else {
		var config Config
		json.Unmarshal(b, &config)
		db := NewMysql(config.DbHost, config.DbPort, config.DbUser, config.DbPwd)
		if err2 := db.Connect(); err2 != nil {
			LogError("connnect db failed, %s", err2.Error())
			return
		}
		qry := NewOrderQueryTimer(db)
		go qry.Start()
		webHandler := NewWebHandlers(db)
		srv := NewWebServer(webHandler, config.ServerListenPort)
		LogInfo("start web server on port : %d", config.ServerListenPort)
		srv.Start()
	}

}
