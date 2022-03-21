// Copyright 2018 ouqiang authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package goproxy

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Context 代理上下文
type Context struct {
	Req         *http.Request
	Data        map[interface{}]interface{}
	TunnelProxy bool
	abort       bool
}

func (c *Context) IsHTTPS() bool {
	return c.Req.URL.Scheme == "https"
}

var defaultPorts = map[string]string{
	"https": "443",
	"http":  "80",
	"":      "80",
}

func (c *Context) WebsocketUrl() *url.URL {
	u := new(url.URL)
	*u = *c.Req.URL
	if c.IsHTTPS() {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	return u
}

func (c *Context) Addr() string {
	addr := c.Req.Host

	if !strings.Contains(c.Req.URL.Host, ":") {
		addr += ":" + defaultPorts[c.Req.URL.Scheme]
	}

	return addr
}

// Abort 中断执行
func (c *Context) Abort() {
	c.abort = true
}

// IsAborted 是否已中断执行
func (c *Context) IsAborted() bool {
	return c.abort
}

// Reset 重置
func (c *Context) Reset(req *http.Request) {
	c.Req = req
	c.Data = make(map[interface{}]interface{})
	c.abort = false
	c.TunnelProxy = false
}

type Delegate interface {
	// Connect 收到客户端连接
	Connect(ctx *Context, rw http.ResponseWriter)
	// Auth 代理身份认证
	Auth(ctx *Context, rw http.ResponseWriter)
	// BeforeRequest HTTP请求前 设置X-Forwarded-For, 修改Header、Body
	BeforeRequest(ctx *Context)
	// BeforeResponse 响应发送到客户端前, 修改Header、Body、Status Code
	BeforeResponse(ctx *Context, resp *http.Response, err error)
	// WebSocketSendMessage websocket发送消息
	WebSocketSendMessage(ctx *Context, messageType *int, p *[]byte)
	// WebSockerReceiveMessage websocket接收 消息
	WebSocketReceiveMessage(ctx *Context, messageType *int, p *[]byte)
	// ParentProxy 上级代理
	ParentProxy(*http.Request) (*url.URL, error)
	// Finish 本次请求结束
	Finish(ctx *Context)
	// 记录错误信息
	ErrorLog(err error)
}

var _ Delegate = &DefaultDelegate{}

// DefaultDelegate 默认Handler什么也不做
type DefaultDelegate struct {
	Delegate
}

func (h *DefaultDelegate) Connect(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) Auth(ctx *Context, rw http.ResponseWriter) {}

func (h *DefaultDelegate) BeforeRequest(ctx *Context) {}

func (h *DefaultDelegate) BeforeResponse(ctx *Context, resp *http.Response, err error) {}

func (h *DefaultDelegate) ParentProxy(req *http.Request) (*url.URL, error) {
	return http.ProxyFromEnvironment(req)
}

// WebSocketSendMessage websocket发送消息
func (h *DefaultDelegate) WebSocketSendMessage(ctx *Context, messageType *int, payload *[]byte) {}

// WebSockerReceiveMessage websocket接收 消息
func (h *DefaultDelegate) WebSocketReceiveMessage(ctx *Context, messageType *int, payload *[]byte) {}

func (h *DefaultDelegate) Finish(ctx *Context) {}

func (h *DefaultDelegate) ErrorLog(err error) {
	log.Println(err)
}
