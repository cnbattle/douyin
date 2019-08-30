package web

import (
	"encoding/json"
	"fmt"
	"github.com/cnbattle/douyin/config"
	"github.com/cnbattle/douyin/database"
	"github.com/cnbattle/douyin/model"
	"github.com/cnbattle/douyin/utils"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func Run() {
	gin.SetMode(config.V.GetString("gin.model"))
	r := gin.Default()
	r.POST("/", handle)

	_ = r.Run(":" + strconv.Itoa(config.V.GetInt("gin.addr")))
}

func handle(ctx *gin.Context) {
	body := ctx.DefaultPostForm("json", "null")
	status := 0
	if body == "null" {
		status = 1
	}

	var data model.Data
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		fmt.Println(err)
	}
	go handleJson(data)
	ctx.JSON(200, gin.H{
		"status":  status,
		"message": "success",
	})
}

func handleJson(data model.Data) {
	for _, item := range data.AwemeList {
		// 判断是否是广告 点赞数是否大于设定值
		if item.IsAds == true || item.Statistics.DiggCount < config.V.GetInt("smallLike") {
			continue
		}
		log.Println("开始处理数据:", item.Desc)

		isDownload := config.V.GetInt("isDownload")
		var localAvatar, localCover, localVideo string
		var err error
		coverUrl, videoUrl := getCoverVideo(&item)
		if isDownload == 1 {
			// 下载封面图 视频 头像图
			localAvatar, localCover, localVideo, err = downloadHttpFile(item.Author.AvatarThumb.UrlList[0], videoUrl, coverUrl)
			if err != nil {
				log.Println("下载封面图 视频 头像图失败:", err)
				continue
			}
		} else {
			localAvatar, localCover, localVideo = item.Author.AvatarThumb.UrlList[0], coverUrl, videoUrl
		}
		// 写入数据库
		var video model.Video
		video.AwemeId = item.AwemeId
		video.Nickname = item.Author.Nickname
		video.Avatar = localAvatar
		video.Desc = item.Desc
		video.CoverPath = localCover
		video.VideoPath = localVideo
		video.IsDownload = isDownload
		database.Local.Create(&video)
	}
}

// downloadHttpFile 下载远程图片
func downloadHttpFile(avatarUrl, videoUrl string, coverUrl string) (string, string, string, error) {
	var localAvatar, localCover, localVideo string
	localAvatar = "download/avatar/" + utils.Md5(avatarUrl) + ".jpeg"
	localVideo = "download/video/" + utils.Md5(videoUrl) + ".mp4"
	localCover = "download/cover/" + utils.Md5(coverUrl) + ".jpeg"
	err := download(avatarUrl, localAvatar)
	if err != nil {
		return "", "", "", err
	}
	err = download(videoUrl, localVideo)
	if err != nil {
		return "", "", "", err
	}
	err = download(coverUrl, localCover)
	if err != nil {
		return "", "", "", err
	}
	return localAvatar, localCover, localVideo, nil
}

// getCoverVideo 获取封面图视频地址
func getCoverVideo(item *model.Item) (coverUrl, videoUrl string) {
	isDownload := config.V.GetInt("isDownload")
	coverUrl = item.Video.Cover.UrlList[0]
	if isDownload == 1 {
		videoUrl = item.Video.PlayAddr.UrlList[0]
		return
	}
	videoUrl = item.Video.PlayAddr.UrlList[len(item.Video.PlayAddr.UrlList)-1]
	return
}

// download 下载文件
func download(url, saveFile string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	res, err := client.Do(req)
	defer res.Body.Close()
	f, err := os.Create(saveFile)
	if err != nil {
		_ = os.Remove(saveFile)
		return err
	}
	_, err = io.Copy(f, res.Body)
	if err != nil {
		_ = os.Remove(saveFile)
		return err
	}
	return nil
}
