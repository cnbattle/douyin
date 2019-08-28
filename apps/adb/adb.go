package adb

import (
	"fmt"
	"os"
	"os/exec"
	"github.com/cnbattle/douyin/config"
	"time"
)

func Run() {
START:
	start := time.Now().Unix()

	runApp()
	defer closeApp()
	for {
		now := time.Now().Unix()
		if now > start+config.V.GetInt64("app.restart") {
			time.Sleep(config.V.GetDuration("app.sleep") * time.Second)
			goto START
		}
		swipe()
		time.Sleep(config.V.GetDuration("swipe.sleep") * time.Millisecond)
	}
}

func runApp() {
	closeApp()
	cmd := exec.Command("./static/adb.exe", "shell", "am", "start", "-n", fmt.Sprintf("%v/%v",
		config.V.GetString("app.packageName"), config.V.GetString("app.startPath"),
	))
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func swipe() {
	cmd := exec.Command("./static/adb.exe", "shell", "input", "swipe",
		config.V.GetString("swipe.startX"),
		config.V.GetString("swipe.startY"),
		config.V.GetString("swipe.endX"),
		config.V.GetString("swipe.endY"),
	)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func closeApp() {
	cmd := exec.Command("./static/adb.exe", "shell", "am", "force-stop", config.V.GetString("app.packageName"))
	cmd.Stdout = os.Stdout
	cmd.Run()
}
