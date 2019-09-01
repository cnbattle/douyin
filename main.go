package main

import (
	"github.com/cnbattle/douyin/internal/adb"
	"github.com/cnbattle/douyin/internal/web"
)

func main() {
	go web.Start()
	go adb.Start()
	select {}
}
