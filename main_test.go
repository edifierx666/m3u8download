package main

import (
	"fmt"
	"m3u8download/utils"
	"os"
	"testing"
)

func Test1(t *testing.T) {
	key := utils.Get("https://ts10.510yh.cc/20210928/O03ECupu/1000kb/hls/key.key").String()
	tsFile := utils.Get("https://ts10.510yh.cc/20210928/O03ECupu/1000kb/hls/lJP0bHhA.ts").Body()
	decrypt, _ := utils.AES128Decrypt(tsFile, []byte(key), []byte(""))
	file, _ := os.Create("/Users/edifierx666/Desktop/53_de.ts")
	file.Write(decrypt)
}
func Test2(t *testing.T) {
	response := utils.Get("https://ts10.510yh.cc/20210928/O03ECupu/1000kb/hls/PxBBa795.ts")
	fmt.Println(response)
}
