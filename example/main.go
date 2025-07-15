package main

import (
	//log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	zero "github.com/cubevlmu/CZeroBot"
	"github.com/cubevlmu/CZeroBot/driver"
	_ "github.com/cubevlmu/CZeroBot/example/command"
	_ "github.com/cubevlmu/CZeroBot/example/music"
	_ "github.com/cubevlmu/CZeroBot/example/napcat"
	_ "github.com/cubevlmu/CZeroBot/example/repeat"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	zero.RunAndBlock(&zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: "/",
		SuperUsers:    []int64{123456},
		Driver: []zero.Driver{
			// 正向 WS
			//driver.NewWebSocketClient("ws://127.0.0.1:6700", ""),
			// 反向 WS
			driver.NewWebSocketServer(16, "ws://127.0.0.1:8080", ""),
			// HTTP
			//driver.NewHTTPClient("http://127.0.0.1:6701", "", "http://127.0.0.1:6700", ""),
		},
	}, nil)
}
