package main

import (
	"github.com/cnbattle/douyin/internal/recommend"
	"github.com/cnbattle/douyin/internal/web"
)

func main() {
	go web.Start()
	go recommend.Start()
	select {}
}
