package adb

import (
	"os"
	"os/exec"
	"runtime"
)

// Command 自定义参数
func Command(arg ...string) {
	cmd := exec.Command(getAdbCli(), arg...)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// RunApp 运行App
func RunApp(startAppPath string) {
	cmd := exec.Command(getAdbCli(), "shell", "am", "start", "-n", startAppPath)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// CloseApp 关闭
func CloseApp(packageName string) {
	cmd := exec.Command(getAdbCli(), "shell", "am", "force-stop", packageName)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// Swipe 滑动
func Swipe(StartX, StartY, EndX, EndY string) {
	cmd := exec.Command(getAdbCli(), "shell", "input", "swipe",
		StartX, StartY, EndX, EndY,
	)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// InputText 输入文本
func InputText(text string) {
	cmd := exec.Command(getAdbCli(), "shell", "input", "text", text)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// InputTextByADBKeyBoard 输入文本
func InputTextByADBKeyBoard(text string) {
	cmd := exec.Command(getAdbCli(), "shell", "am", "broadcast", "-a", "ADB_INPUT_TEXT", "--es", "msg", text)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// Click 点击某一像素点
func Click(X, Y string) {
	cmd := exec.Command(getAdbCli(), "shell", "input", "tap", X, Y)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// ClickKeyCode 点击android对应的keycode
func ClickKeyCode(code string) {
	cmd := exec.Command(getAdbCli(), "shell", "input", "keyevent", code)
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}

// ClickHome 点击hone键
func ClickHome() {
	ClickKeyCode("3")
}

// ClickHome 点击返回键
func ClickBack() {
	ClickKeyCode("4")
}

// getAdbCli 获取adb cli
func getAdbCli() string {
	if runtime.GOOS == "windows" {
		return "./static/adb/adb.exe"
	}
	return "adb"
}
