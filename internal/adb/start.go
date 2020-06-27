package adb

import (
	"github.com/cnbattle/douyin/internal/config"
	"log"
	"time"
)

func init() {
	log.Println("[ADB] start")
}

func Start() {
START:
	start := time.Now().Unix()
	CloseApp(config.V.GetString("app.packageName"))
	RunApp(config.V.GetString("app.packageName") + "/" + config.V.GetString("app.startPath"))
	time.Sleep(10 * time.Second)
	for {
		now := time.Now().Unix()
		if now > start+config.V.GetInt64("app.restart") {
			time.Sleep(config.V.GetDuration("app.sleep") * time.Second)
			goto START
		}
		Click(config.V.GetString("click.x"), config.V.GetString("click.y"))
		time.Sleep(300 * time.Millisecond)
		Click(config.V.GetString("click.x"), config.V.GetString("click.y"))

		time.Sleep(config.V.GetDuration("click.sleep") * time.Second)
	}
}
