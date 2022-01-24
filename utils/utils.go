package utils

import (
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"net/url"
	"os"
	"path"
)

type EnvHubs struct {
	Curpath string
}

var EnvhubStore *EnvHubs

func init() {

	EnvhubStore = &EnvHubs{
		Curpath: "",
	}
	setCurpath()
}

func setCurpath() {
	curdirpath, err := os.Executable()
	if err != nil {
		EnvhubStore.Curpath, _ = os.Getwd()
	}
	EnvhubStore.Curpath = path.Dir(curdirpath)
}

var FakeHeaders = map[string]string{
	// "Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	// "Accept-Charset":  "UTF-8,*;q=0.5",
	// "Accept-Encoding": "gzip,deflate,sdch",
	// "Accept-Language": "en-US,en;q=0.8",
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 Safari/537.36",
}

var defRequest = resty.New()

var Fetch = defRequest

func init() {
	defRequest.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})
	// defRequest.SetHeaders(FakeHeaders).
	// 	SetRetryCount(3).
	// 	SetRetryWaitTime(5 * time.Second).
	// 	SetRetryMaxWaitTime(20 * time.Second)
}
func ConcatDomain(u string, tar string) string {
	tarUrl, err := url.Parse(tar)
	if err != nil {
		Logger.Fatal("添加domain,url错误:", zap.String("u", u))
	}
	uUrl, _ := url.Parse(u)
	if tarUrl.Hostname() == "" {
		return uUrl.ResolveReference(tarUrl).String()
	}
	return tar
}

func Get(u string) *resty.Response {
	resp, err := Fetch.R().Get(u)
	if err != nil {
		Logger.Fatal("GET获取URL出错", zap.String("url", u), zap.String("error", err.Error()))
	}
	return resp
}
