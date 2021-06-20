package utils

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io/ioutil"
	"os"
)

// Md5 字符串加密
func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// ParseGzip gzip 解压
func ParseGzip(data []byte) ([]byte, error) {
	b := new(bytes.Buffer)
	binary.Write(b, binary.LittleEndian, data)
	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	} else {
		defer r.Close()
		undatas, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return undatas, nil
	}
}
