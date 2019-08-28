package main

import (
	"github.com/cnbattle/douyin/apps/web"
)

func main() {
	go web.Run()
	//go adb.Run()
	select {}
}
