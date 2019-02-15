package main

import (
	"fmt"
	"strings"
	"bytes"
	"reflect"
	"net/http"
	"io/ioutil"
	"golang.org/x/net/context"
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"github.com/akkuman/parseConfig"
)

var(
	accessKey, secretKey, bucket, upToken, key, videoDomain string
	cfg storage.Config
	config parseConfig.Config
)

type Qiniu struct {
	accessKey string
	secretKey string
	bucket string
	token string
	domain string
	uploader *storage.FormUploader
}

func (qiniu * Qiniu) init() {
	config = parseConfig.New("./config.json")
	qiniu.accessKey = config.Get("qiniu_config > accessKey").(string)
	qiniu.secretKey = config.Get("qiniu_config > secretKey").(string)
	qiniu.bucket = config.Get("qiniu_config > bucket").(string)
	qiniu.domain =  config.Get("qiniu_config > url").(string)

	cfg := storage.Config{}
	cfg.Zone = &storage.ZoneHuadong
	cfg.UseHTTPS = false
	cfg.UseCdnDomains = false
	qiniu.uploader = storage.NewFormUploader(&cfg)
}

func (qiniu * Qiniu) refreshToken() {
	mac := qbox.NewMac(qiniu.accessKey, qiniu.secretKey)
	putPolicy := storage.PutPolicy{
		Scope: qiniu.bucket,
	}
	qiniu.token = putPolicy.UploadToken(mac)
}

func (qiniu * Qiniu) uploadRemoteFile(url string) {
	qiniu.refreshToken()
	s := strings.Split(url, "/")
	key := s[len(s) - 1]
	//将url文件读取到本地内存,然后使用字节数组上传的方式
	res,err := http.Get(url)
	if err != nil || res.StatusCode != 200{
		fmt.Printf("下载失败:%s", res.Request.URL)
	}
	data,err2 := ioutil.ReadAll(res.Body)
	fmt.Println("type:", reflect.TypeOf(data))
	if err2 != nil {
		fmt.Printf("读取数据失败")
	}
	dataLen := int64(len(data))
	ret := storage.PutRet{}
	err3 := qiniu.uploader.Put(context.Background(), &ret, qiniu.token, key, bytes.NewReader(data), dataLen, nil)
	if err3 != nil {
		fmt.Println(err3)
		return
	}
	fmt.Println(ret.Key, ret.Hash)
}

func (qiniu * Qiniu) uploadLocalFile(file string) {
	//todo token重用
	qiniu.refreshToken()
	key := file
	ret := storage.PutRet{}
	err := qiniu.uploader.PutFile(context.Background(), &ret, qiniu.token, key, file, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret.Key,ret.Hash)
	// return ret, err
}

func main() {
	// file := "test.jpg"
	// var q *Qiniu = new(Qiniu)
	// q.init()
	// q.uploadLocalFile(file)

	path := "https://dev-oss.secon.cn/pic_pingtaizhibo.png"
	var q *Qiniu = new(Qiniu)
	q.init()
	q.uploadRemoteFile(path)
}
