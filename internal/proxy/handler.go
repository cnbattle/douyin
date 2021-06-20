package proxy

import (
	"bytes"
	"encoding/json"
	"github.com/ouqiang/goproxy"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/cnbattle/douyin/internal/core"
	"github.com/cnbattle/douyin/internal/database/model"
	"github.com/cnbattle/douyin/internal/utils"
)

type EventHandler struct{}

func (e *EventHandler) Connect(ctx *goproxy.Context, rw http.ResponseWriter) {
	// 保存的数据可以在后面的回调方法中获取
	ctx.Data["req_id"] = "uuid"

	// 禁止访问某个域名
	if strings.Contains(ctx.Req.URL.Host, "example.com") {
		rw.WriteHeader(http.StatusForbidden)
		ctx.Abort()
		return
	}
}

func (e *EventHandler) Auth(ctx *goproxy.Context, rw http.ResponseWriter) {
	// 身份验证
}

func (e *EventHandler) BeforeRequest(ctx *goproxy.Context) {
	// 修改header
	ctx.Req.Header.Add("X-Request-Id", ctx.Data["req_id"].(string))
	// 设置X-Forwarded-For
	if clientIP, _, err := net.SplitHostPort(ctx.Req.RemoteAddr); err == nil {
		if prior, ok := ctx.Req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		ctx.Req.Header.Set("X-Forwarded-For", clientIP)
	}
	// 读取Body
	body, err := ioutil.ReadAll(ctx.Req.Body)
	if err != nil {
		// 错误处理
		return
	}
	// Request.Body只能读取一次, 读取后必须再放回去
	// Response.Body同理
	ctx.Req.Body = ioutil.NopCloser(bytes.NewReader(body))

}

func (e *EventHandler) BeforeResponse(ctx *goproxy.Context, resp *http.Response, err error) {
	if err != nil {
		return
	}
	// /aweme/v1/general/search/single/  综合搜索
	// /aweme/v1/search/item/ 视频
	//if strings.EqualFold(ctx.Req.URL.Path, "/aweme/v1/general/search/single/") {
	//	response, err := ioutil.ReadAll(resp.Body)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	// gzip
	//	body, err := utils.ParseGzip(response)
	//	if err != nil {
	//		log.Println(err)
	//		return
	//	}
	//	var filename = "./single.json"
	//	var f *os.File
	//	/***************************** 第一种方式: 使用 io.WriteString 写入文件 ***********************************************/
	//	if utils.CheckFileIsExist(filename) { //如果文件存在
	//		f, _ = os.OpenFile(filename, os.O_APPEND, 0666) //打开文件
	//		fmt.Println("文件存在")
	//	} else {
	//		f, _ = os.Create(filename) //创建文件
	//		fmt.Println("文件不存在")
	//	}
	//	_, _ = io.WriteString(f, string(body)) //写入文件(字符串)
	//	//var data model.Data
	//	//err = json.Unmarshal(body, &data)
	//	//if err != nil {
	//	//	log.Println(err)
	//	//	return
	//	//}
	//	//go core.HandleJson(data)
	//	// resp.Body 只能读取一次, 读取后必须再放回去
	//	resp.Body = ioutil.NopCloser(bytes.NewReader(response))
	//}

	// 处理 推荐列表接口  搜索页视频列表接口
	if strings.EqualFold(ctx.Req.URL.Path, "/aweme/v1/feed/") || strings.EqualFold(ctx.Req.URL.Path, "/aweme/v1/search/item/") {
		response, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}
		// gzip
		body, err := utils.ParseGzip(response)
		if err != nil {
			log.Println(err)
			return
		}
		var data model.Data
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Println(err)
			return
		}
		go core.HandleJson(data)
		// resp.Body 只能读取一次, 读取后必须再放回去
		resp.Body = ioutil.NopCloser(bytes.NewReader(response))
	}
}

// 设置上级代理
func (e *EventHandler) ParentProxy(req *http.Request) (*url.URL, error) {
	return nil, nil
}

// Finish 请求结束
func (e *EventHandler) Finish(ctx *goproxy.Context) {

}

// ErrorLog 记录错误日志
func (e *EventHandler) ErrorLog(err error) {

}
