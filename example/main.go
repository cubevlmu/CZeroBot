package main

import (
	"flag"
	"github.com/wdvxdr1123/ZeroBot/extension/shell"
	"github.com/wdvxdr1123/ZeroBot/message"

	//log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	_ "github.com/wdvxdr1123/ZeroBot/example/command"
	_ "github.com/wdvxdr1123/ZeroBot/example/music"
	_ "github.com/wdvxdr1123/ZeroBot/example/napcat"
	_ "github.com/wdvxdr1123/ZeroBot/example/repeat"
)

func init() {
	log.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zero][%time%][%lvl%]: %msg% \n",
	})
	log.SetLevel(log.DebugLevel)
}

func main() {
	zero.OnCommand("github").Handle(func(ctx *zero.Ctx) {
		fset := flag.FlagSet{}
		var (
			owner string
			repo  string
		)
		fset.StringVar(&owner, "o", "wdvxdr1123", "")
		fset.StringVar(&repo, "r", "ZeroBot", "")
		arguments := shell.Parse(ctx.State["args"].(string))
		err := fset.Parse(arguments)
		if err != nil {
			return
		}
		ctx.Send(message.Text("github\n" +
			"owner: " + owner + "\n" +
			"repo: " + repo,
		))
	})
	
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
