package proxy

import (
	"log"
	"net/http"
	"time"

	"github.com/cnbattle/douyin/internal/config"

	"github.com/ouqiang/goproxy"
)

var addr = ":8080"

func init() {
	addr = config.V.GetString("proxy.addr")
	log.Println("[Proxy] Listen:", addr)
}

func Start() {
	proxy := goproxy.New(goproxy.WithDelegate(&EventHandler{}), goproxy.WithDecryptHTTPS(&Cache{}))
	server := &http.Server{
		Addr:         addr,
		Handler:      proxy,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
