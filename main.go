package main

import (
	"m3u8download/m3u8parser"
	"m3u8download/options"
)

func main() {
	// https://sod.bunediy.com/20211208/Ix3uVqFY/index.m3u8
	// https://v10.dious.cc/20210928/O03ECupu/index.m3u8
	// https://bitdash-a.akamaihd.net/content/sintel/hls/video/1500kbit.m3u8
	// https://vod1.bdzybf1.com/20200813/z7wkUMUm/index.m3u8
	// https://s1.monidai.com/20211227/7A6XZbXe/index.m3u8
	o := options.New("https://s1.monidai.com/20211227/7A6XZbXe/index.m3u8")
	o.Output = "/Users/edifierx666/Desktop"
	m3u8parser.New(o).Parse()
}
