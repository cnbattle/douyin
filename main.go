package main

import (
	"github.com/cnbattle/douyin/internal/adb"
	"github.com/cnbattle/douyin/internal/proxy"
)

func main() {
	go proxy.Start()
	go adb.Start()
	select {}
}
