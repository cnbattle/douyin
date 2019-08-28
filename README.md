# 抖音推荐列表视频爬虫方案

基于APP爬取

技术栈：golang adb nodejs anyproxy

![](example.gif)

## 使用

1. 按照anyproxy, 详细请自己google

1. 使用android虚拟机或使用真机,安装抖音 ,配置anyproxy https代理

1. 修改anyproxy 配置文件,详见 angproxy目录下文件,具体看`beforeSendRequest` `beforeSendResponse` 函数代码

1. 启动anyproxy(用pm2启动管理最佳)

1. 复制 `config.example.toml` 为 `config.toml`,并根据自己需求修改参数

1. 运行 本项目程序 `go run main.go` 或 编辑运行也可

1. 最后会生成一个 `database.db`的sqlite3数据库文件,字符详见`model/videos.go`文件,静态文件(用户头像,视频封面图,视频文件)将放在`download/[avatar,cover,video]`目录下

## 最后说明

- 个人能力一般,有很多编码不规范的地方请包涵
- 有能力的朋友可以根据个人需求修改,如果可以请提交pr
- 如果使用有问题,请提交`issues` 或加我同名微信号,请备注github过来的,谢谢
