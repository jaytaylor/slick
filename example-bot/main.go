package main

import (
	"flag"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/abourget/slick"
	_ "github.com/abourget/slick/bugger"
	_ "github.com/abourget/slick/deployer"
	_ "github.com/abourget/slick/faceoff"
	_ "github.com/abourget/slick/funny"
	_ "github.com/abourget/slick/healthy"
	_ "github.com/abourget/slick/hooker"
	_ "github.com/abourget/slick/mooder"
	_ "github.com/abourget/slick/plotberry"
	_ "github.com/abourget/slick/recognition"
	_ "github.com/abourget/slick/standup"
	_ "github.com/abourget/slick/todo"
	_ "github.com/abourget/slick/totw"
	_ "github.com/abourget/slick/web"
	_ "github.com/abourget/slick/webauth"
	_ "github.com/abourget/slick/webutils"
	_ "github.com/abourget/slick/wicked"
)

var configFile = flag.String("config", os.Getenv("HOME")+"/.slick.conf", "config file")

func main() {
	log.SetLevel(log.DebugLevel)

	flag.Parse()

	bot := slick.New(*configFile)

	bot.Run()
}
