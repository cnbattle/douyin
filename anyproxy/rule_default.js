'use strict';

module.exports = {

    summary: 'the default rule for AnyProxy',

    /**
     *
     *
     * @param {object} requestDetail
     * @param {string} requestDetail.protocol
     * @param {object} requestDetail.requestOptions
     * @param {object} requestDetail.requestData
     * @param {object} requestDetail.response
     * @param {number} requestDetail.response.statusCode
     * @param {object} requestDetail.response.header
     * @param {buffer} requestDetail.response.body
     * @returns
     */
    * beforeSendRequest(requestDetail) {
        const localResponse = {
            statusCode: 200,
            header: {
                'Content-Type': 'image/gif'
            },
            body: 'data:image/gif;base64,R0lGODlhAQABAIAAAP///wAAACH5BAEAAAAALAAAAAABAAEAAAICRAEAOw=='
        };
        if (/byteimg\.com/i.test(requestDetail.url)) { //图片链接
            return {
                response: localResponse
            }
        }        
        if (/ixigua\.com/i.test(requestDetail.url)) { //视频链接
            return {
                response: localResponse
            }
        }
        
        if (/google/i.test(requestDetail.url)) {
            return {
                response: {
                    statusCode: 200,
                    header: {
                        'Content-Type': 'application/json'
                    },
                    body: '[]'
                }
            }
        }
        return null;
    },


    /**
     *
     *
     * @param {object} requestDetail
     * @param {object} responseDetail
     */
    * beforeSendResponse(requestDetail, responseDetail) {
        if(/aweme\/v1\/feed/i.test(requestDetail.url)){
            var json = responseDetail.response.body.toString()
            //将匹配到的json发送到自己的服务器
            HttpPost({json: json}, "/");
        }
        return null;
    },


    /**
     * default to return null
     * the user MUST return a boolean when they do implement the interface in rule
     *
     * @param {any} requestDetail
     * @returns
     */
    * beforeDealHttpsRequest(requestDetail) {
        if (requestDetail.replaceLocalFile) {
            callback(200, {
                "content-type": "image/png"
            }, img);
        }
        return null;
    },

    /**
     *
     *
     * @param {any} requestDetail
     * @param {any} error
     * @returns
     */
    * onError(requestDetail, error) {
        return null;
    },


    /**
     *
     *
     * @param {any} requestDetail
     * @param {any} error
     * @returns
     */
    * onConnectError(requestDetail, error) {
        return null;
    },
};


//将json发送到服务器，str为json内容，url为历史消息页面地址，path是接收程序的路径和文件名
function HttpPost(data, path) {
    var http_url = 'http://127.0.0.1:8080' + path;
    var content = require('querystring').stringify(data);
    var parse_u = require('url').parse(http_url, true);
    var isHttp = parse_u.protocol == 'http:';
    var options = {
        host: parse_u.hostname,
        port: parse_u.port || (isHttp ? 80 : 443),
        path: parse_u.path,
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Content-Length': content.length
        }
    };
    var req = require(isHttp ? 'http' : 'https').request(options, function (res) {
        var _data = '';
        res.on('data', function (chunk) {
            _data += chunk;
        });
        res.on('end', function () {
            // console.log("\n--->>\nresult:", _data)
        });
    });
    req.write(content);
    req.end();
}