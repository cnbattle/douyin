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

// Package goproxy HTTP(S)代理, 支持中间人代理解密HTTPS数据
package goproxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ouqiang/websocket"

	"github.com/ouqiang/goproxy/cert"
)

const (
	// 连接目标服务器超时时间
	defaultTargetConnectTimeout = 5 * time.Second
	// 目标服务器读写超时时间
	defaultTargetReadWriteTimeout = 30 * time.Second
	// 客户端读写超时时间
	defaultClientReadWriteTimeout = 30 * time.Second
)

// 隧道连接成功响应行
var tunnelEstablishedResponseLine = []byte("HTTP/1.1 200 Connection established\r\n\r\n")

var badGateway = []byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n\r\n", http.StatusBadGateway, http.StatusText(http.StatusBadGateway)))

var (
	bufPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 32*1024)
		},
	}

	ctxPool = sync.Pool{
		New: func() interface{} {
			return new(Context)
		},
	}
	headerPool  = NewHeaderPool()
	requestPool = newRequestPool()
)

type RequestPool struct {
	pool sync.Pool
}

func newRequestPool() *RequestPool {
	return &RequestPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(http.Request)
			},
		},
	}
}

func (p *RequestPool) Get() *http.Request {
	req := p.pool.Get().(*http.Request)

	req.Method = ""
	req.URL = nil
	req.Proto = ""
	req.ProtoMajor = 0
	req.ProtoMinor = 0
	req.Header = nil
	req.Body = nil
	req.GetBody = nil
	req.ContentLength = 0
	req.TransferEncoding = nil
	req.Close = false
	req.Host = ""
	req.Form = nil
	req.PostForm = nil
	req.MultipartForm = nil
	req.Trailer = nil
	req.RemoteAddr = ""
	req.RequestURI = ""
	req.TLS = nil
	req.Cancel = nil
	req.Response = nil

	return req
}

func (p *RequestPool) Put(req *http.Request) {
	if req != nil {
		p.pool.Put(req)
	}
}

type HeaderPool struct {
	pool sync.Pool
}

func NewHeaderPool() *HeaderPool {
	return &HeaderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return http.Header{}
			},
		},
	}
}

func (p *HeaderPool) Get() http.Header {
	header := p.pool.Get().(http.Header)
	for k := range header {
		delete(header, k)
	}

	return header
}

func (p *HeaderPool) Put(header http.Header) {
	if header != nil {
		p.pool.Put(header)
	}
}

// 生成隧道建立请求行
func makeTunnelRequestLine(addr string) string {
	return fmt.Sprintf("CONNECT %s HTTP/1.1\r\n\r\n", addr)
}

type options struct {
	disableKeepAlive bool
	delegate         Delegate

	decryptHTTPS       bool
	websocketIntercept bool
	certCache          cert.Cache
	transport          *http.Transport
}

type Option func(*options)

// WithDisableKeepAlive 连接是否重用
func WithDisableKeepAlive(disableKeepAlive bool) Option {
	return func(opt *options) {
		opt.disableKeepAlive = disableKeepAlive
	}
}

// WithDelegate 设置委托类
func WithDelegate(delegate Delegate) Option {
	return func(opt *options) {
		opt.delegate = delegate
	}
}

// WithTransport 自定义http transport
func WithTransport(t *http.Transport) Option {
	return func(opt *options) {
		opt.transport = t
	}
}

// WithDecryptHTTPS 中间人代理, 解密HTTPS, 需实现证书缓存接口
func WithDecryptHTTPS(c cert.Cache) Option {
	return func(opt *options) {
		opt.decryptHTTPS = true
		opt.certCache = c
	}
}

// WithEnableWebsocketIntercept 拦截websocket
func WithEnableWebsocketIntercept() Option {
	return func(opt *options) {
		opt.websocketIntercept = true
	}
}

// New 创建proxy实例
func New(opt ...Option) *Proxy {
	opts := &options{}
	for _, o := range opt {
		o(opts)
	}
	if opts.delegate == nil {
		opts.delegate = &DefaultDelegate{}
	}
	if opts.transport == nil {
		opts.transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	}

	p := &Proxy{}
	p.delegate = opts.delegate
	p.websocketIntercept = opts.websocketIntercept
	p.decryptHTTPS = opts.decryptHTTPS
	if p.decryptHTTPS {
		p.cert = cert.NewCertificate(opts.certCache)
	}
	p.transport = opts.transport
	p.transport.DisableKeepAlives = opts.disableKeepAlive
	p.transport.Proxy = p.delegate.ParentProxy

	return p
}

// Proxy 实现了http.Handler接口
type Proxy struct {
	delegate           Delegate
	clientConnNum      int32
	decryptHTTPS       bool
	websocketIntercept bool
	cert               *cert.Certificate
	transport          *http.Transport
}

var _ http.Handler = &Proxy{}

// ServeHTTP 实现了http.Handler接口
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Host == "" {
		req.URL.Host = req.Host
	}
	atomic.AddInt32(&p.clientConnNum, 1)
	ctx := ctxPool.Get().(*Context)
	ctx.Reset(req)

	defer func() {
		p.delegate.Finish(ctx)
		ctxPool.Put(ctx)
		atomic.AddInt32(&p.clientConnNum, -1)
	}()
	p.delegate.Connect(ctx, rw)
	if ctx.abort {
		return
	}
	p.delegate.Auth(ctx, rw)
	if ctx.abort {
		return
	}

	switch {
	case ctx.Req.Method == http.MethodConnect:
		p.tunnelProxy(ctx, rw)
	case websocket.IsWebSocketUpgrade(ctx.Req):
		p.tunnelProxy(ctx, rw)
	default:
		p.httpProxy(ctx, rw)
	}
}

// ClientConnNum 获取客户端连接数
func (p *Proxy) ClientConnNum() int32 {
	return atomic.LoadInt32(&p.clientConnNum)
}

// DoRequest 执行HTTP请求，并调用responseFunc处理response
func (p *Proxy) DoRequest(ctx *Context, responseFunc func(*http.Response, error)) {
	if ctx.Data == nil {
		ctx.Data = make(map[interface{}]interface{})
	}
	p.delegate.BeforeRequest(ctx)
	if ctx.abort {
		return
	}
	newReq := requestPool.Get()
	*newReq = *ctx.Req
	newHeader := headerPool.Get()
	CloneHeader(newReq.Header, newHeader)
	newReq.Header = newHeader
	for _, item := range hopHeaders {
		if newReq.Header.Get(item) != "" {
			newReq.Header.Del(item)
		}
	}
	resp, err := p.transport.RoundTrip(newReq)
	p.delegate.BeforeResponse(ctx, resp, err)
	if ctx.abort {
		return
	}
	if err == nil {
		for _, h := range hopHeaders {
			resp.Header.Del(h)
		}
	}
	responseFunc(resp, err)
	headerPool.Put(newHeader)
	requestPool.Put(newReq)
}

// HTTP代理
func (p *Proxy) httpProxy(ctx *Context, rw http.ResponseWriter) {
	ctx.Req.URL.Scheme = "http"
	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTP请求错误: %s", ctx.Req.URL, err))
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
		defer func() {
			_ = resp.Body.Close()
		}()
		CopyHeader(rw.Header(), resp.Header)
		rw.WriteHeader(resp.StatusCode)
		buf := bufPool.Get().([]byte)
		_, _ = io.CopyBuffer(rw, resp.Body, buf)
		bufPool.Put(buf)
	})
}

// HTTPS代理
func (p *Proxy) httpsProxy(ctx *Context, clientConn net.Conn) {
	tlsConfig, err := p.cert.GenerateTlsConfig(ctx.Req.URL.Host)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 生成证书失败: %s", ctx.Req.URL.Host, err))
		return
	}
	tlsClientConn := tls.Server(clientConn, tlsConfig)
	_ = tlsClientConn.SetDeadline(time.Now().Add(defaultClientReadWriteTimeout))
	defer func() {
		_ = tlsClientConn.Close()
	}()
	if err := tlsClientConn.Handshake(); err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 握手失败: %s", ctx.Req.URL.Host, err))
		return
	}
	_ = tlsClientConn.SetDeadline(time.Time{})

	buf := bufio.NewReader(tlsClientConn)
	tlsReq, err := http.ReadRequest(buf)
	if err != nil {
		if err != io.EOF {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 读取客户端请求失败: %s", ctx.Req.URL.Host, err))
		}
		return
	}
	tlsReq.RemoteAddr = ctx.Req.RemoteAddr
	tlsReq.URL.Scheme = "https"
	tlsReq.URL.Host = tlsReq.Host

	ctx.Req = tlsReq
	if websocket.IsWebSocketUpgrade(ctx.Req) {
		p.websocketProxy(ctx, NewConnBuffer(tlsClientConn, nil))
		return
	}

	p.DoRequest(ctx, func(resp *http.Response, err error) {
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, 请求错误: %s", ctx.Req.URL, err))
			_, _ = tlsClientConn.Write(badGateway)
			return
		}
		err = resp.Write(tlsClientConn)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - HTTPS解密, response写入客户端失败, %s", ctx.Req.URL, err))
		}
		_ = resp.Body.Close()
	})
}

// 隧道代理
func (p *Proxy) tunnelProxy(ctx *Context, rw http.ResponseWriter) {
	clientConn, err := hijacker(rw)
	if err != nil {
		p.delegate.ErrorLog(err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() {
		_ = clientConn.Close()
	}()

	if websocket.IsWebSocketUpgrade(ctx.Req) {
		p.websocketProxy(ctx, clientConn)
		return
	}

	parentProxyURL, err := p.delegate.ParentProxy(ctx.Req)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - 解析代理地址错误: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	if parentProxyURL == nil {
		_, err = clientConn.Write(tunnelEstablishedResponseLine)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - 隧道连接成功,通知客户端错误: %s", ctx.Req.URL.Host, err))
			return
		}
	}

	isWebsocket := p.detectConnProtocol(clientConn)
	if isWebsocket {
		req, err := http.ReadRequest(clientConn.BufferReader())
		if err != nil {
			if err != io.EOF {
				p.delegate.ErrorLog(fmt.Errorf("%s - websocket读取客户端升级请求失败: %s", ctx.Req.URL.Host, err))
			}
			return
		}
		req.RemoteAddr = ctx.Req.RemoteAddr
		req.URL.Scheme = "http"
		req.URL.Host = req.Host
		ctx.Req = req

		p.websocketProxy(ctx, clientConn)
		return
	}

	targetAddr := ctx.Req.URL.Host
	if parentProxyURL != nil {
		targetAddr = parentProxyURL.Host
	}

	targetConn, err := net.DialTimeout("tcp", targetAddr, defaultTargetConnectTimeout)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - 隧道转发连接目标服务器失败: %s", ctx.Req.URL.Host, err))
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer func() {
		_ = targetConn.Close()
	}()
	_ = clientConn.SetDeadline(time.Now().Add(defaultClientReadWriteTimeout))
	_ = targetConn.SetDeadline(time.Now().Add(defaultTargetReadWriteTimeout))
	if parentProxyURL != nil {
		tunnelRequestLine := makeTunnelRequestLine(ctx.Req.URL.Host)
		_, _ = targetConn.Write([]byte(tunnelRequestLine))
	}

	if p.decryptHTTPS {
		p.httpsProxy(ctx, clientConn)
	} else {
		p.tunnelConnected(ctx)
		p.transfer(clientConn, targetConn)
	}
}

// WebSocket代理
func (p *Proxy) websocketProxy(ctx *Context, srcConn *ConnBuffer) {
	p.tunnelConnected(ctx)

	if !p.websocketIntercept {
		remoteAddr := ctx.Addr()
		var err error
		var targetConn net.Conn
		if ctx.IsHTTPS() {
			targetConn, err = tls.Dial("tcp", remoteAddr, &tls.Config{InsecureSkipVerify: true})
		} else {
			targetConn, err = net.Dial("tcp", remoteAddr)
		}
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - websocket连接目标服务器错误: %s", ctx.Req.URL.Host, err))
			return
		}
		err = ctx.Req.Write(targetConn)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("%s - websocket协议转换请求写入目标服务器错误: %s", ctx.Req.URL.Host, err))
			return
		}
		p.transfer(srcConn, targetConn)
		return
	}

	up := &websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	srcWSConn, err := up.Upgrade(srcConn, ctx.Req, http.Header{})
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - 源连接升级到websocket协议错误: %s", ctx.Req.URL.Host, err))
		return
	}

	u := ctx.WebsocketUrl()
	d := websocket.Dialer{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	targetWSConn, _, err := d.Dial(u.String(), ctx.Req.Header)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("%s - 目标连接升级到websocket协议错误: %s", ctx.Req.URL.Host, err))
		return
	}
	p.transferWebsocket(ctx, srcWSConn, targetWSConn)
}

// 探测连接协议
func (p *Proxy) detectConnProtocol(connBuf *ConnBuffer) (isWebsocket bool) {
	methodBytes, err := connBuf.Peek(3)
	if err != nil {
		return false
	}
	method := string(methodBytes)
	if method != http.MethodGet {
		return false
	}

	return true
}

// webSocket双向转发
func (p *Proxy) transferWebsocket(ctx *Context, srcConn *websocket.Conn, targetConn *websocket.Conn) {
	doneCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for {
			if doneCtx.Err() != nil {
				return
			}

			msgType, msg, err := srcConn.ReadMessage()
			if err != nil {
				p.delegate.ErrorLog(fmt.Errorf("websocket消息转发错误: [%s -> %s] %s", srcConn.RemoteAddr(), targetConn.RemoteAddr(), err))
				return
			}
			p.delegate.WebSocketSendMessage(ctx, &msgType, &msg)
			err = targetConn.WriteMessage(msgType, msg)
			if err != nil {
				p.delegate.ErrorLog(fmt.Errorf("websocket消息转发错误: [%s -> %s] %s", srcConn.RemoteAddr(), targetConn.RemoteAddr(), err))
				return
			}
		}
	}()

	for {
		if doneCtx.Err() != nil {
			return
		}

		msgType, msg, err := targetConn.ReadMessage()
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("websocket消息转发错误: [%s -> %s] %s", targetConn.RemoteAddr(), srcConn.RemoteAddr(), err))
			return
		}
		p.delegate.WebSocketReceiveMessage(ctx, &msgType, &msg)
		err = srcConn.WriteMessage(msgType, msg)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("websocket消息转发错误: [%s -> %s] %s", targetConn.RemoteAddr(), srcConn.RemoteAddr(), err))
			return
		}
	}
}

// 双向转发
func (p *Proxy) transfer(src net.Conn, dst net.Conn) {
	go func() {
		buf := bufPool.Get().([]byte)
		_, err := io.CopyBuffer(src, dst, buf)
		if err != nil {
			p.delegate.ErrorLog(fmt.Errorf("隧道双向转发错误: [%s -> %s] %s", dst.RemoteAddr().String(), src.RemoteAddr().String(), err))
		}
		bufPool.Put(buf)
		_ = src.Close()
		_ = dst.Close()
	}()

	buf := bufPool.Get().([]byte)
	_, err := io.CopyBuffer(dst, src, buf)
	if err != nil {
		p.delegate.ErrorLog(fmt.Errorf("隧道双向转发错误: [%s -> %s] %s", src.RemoteAddr().String(), dst.RemoteAddr().String(), err))
	}
	bufPool.Put(buf)
	_ = dst.Close()
	_ = src.Close()
}

func (p *Proxy) tunnelConnected(ctx *Context) {
	ctx.TunnelProxy = true
	p.delegate.BeforeRequest(ctx)
	resp := &http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	p.delegate.BeforeResponse(ctx, resp, nil)
}

// 获取底层连接
func hijacker(rw http.ResponseWriter) (*ConnBuffer, error) {
	hijacker, ok := rw.(http.Hijacker)
	if !ok {
		return nil, fmt.Errorf("http server不支持Hijacker")
	}
	conn, buf, err := hijacker.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijacker错误: %s", err)
	}

	return NewConnBuffer(conn, buf), nil
}

// CopyHeader 浅拷贝Header
func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// CloneHeader 深拷贝Header
func CloneHeader(h http.Header, h2 http.Header) {
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
}

// CloneBody 拷贝Body
func CloneBody(b io.ReadCloser) (r io.ReadCloser, body []byte, err error) {
	if b == nil {
		return http.NoBody, nil, nil
	}
	body, err = ioutil.ReadAll(b)
	if err != nil {
		return http.NoBody, nil, err
	}
	r = ioutil.NopCloser(bytes.NewReader(body))

	return r, body, nil
}

var hopHeaders = []string{
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
}

type ConnBuffer struct {
	net.Conn
	buf *bufio.ReadWriter
}

func NewConnBuffer(conn net.Conn, buf *bufio.ReadWriter) *ConnBuffer {
	if buf == nil {
		buf = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	}
	return &ConnBuffer{
		Conn: conn,
		buf:  buf,
	}
}

func (cb *ConnBuffer) BufferReader() *bufio.Reader {
	return cb.buf.Reader
}

func (cb *ConnBuffer) Read(b []byte) (n int, err error) {
	return cb.buf.Read(b)
}

func (cb *ConnBuffer) Peek(n int) ([]byte, error) {
	return cb.buf.Peek(n)
}

func (cb *ConnBuffer) Write(p []byte) (n int, err error) {
	n, err = cb.buf.Write(p)
	if err != nil {
		return 0, err
	}

	return n, cb.buf.Flush()
}

func (cb *ConnBuffer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return cb.Conn, cb.buf, nil
}

func (cb *ConnBuffer) WriteHeader(_ int) {}

func (cb *ConnBuffer) Header() http.Header { return nil }
