package recommend

import (
	"github.com/cnbattle/douyin/config"
	"github.com/cnbattle/douyin/internal/adb"
	"time"
)

func Start() {
START:
	start := time.Now().Unix()
	adb.CloseApp(config.V.GetString("app.packageName"))
	adb.RunApp(config.V.GetString("app.packageName") + "/" + config.V.GetString("app.startPath"))
	for {
		now := time.Now().Unix()
		if now > start+config.V.GetInt64("app.restart") {
			time.Sleep(config.V.GetDuration("app.sleep") * time.Second)
			goto START
		}
		adb.Swipe(config.V.GetString("swipe.startX"), config.V.GetString("swipe.startY"),
			config.V.GetString("swipe.endX"), config.V.GetString("swipe.endY"))
		time.Sleep(config.V.GetDuration("swipe.sleep") * time.Millisecond)
	}
}
