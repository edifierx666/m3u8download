package options

import (
	"os"
	"runtime"
)

type Options struct {
	U                      string
	Output                 string
	Cookies                string
	Debug                  bool
	MaxDownloadCount       int
	PerRequestIntervalTime int
}

func New(u string) *Options {
	return &Options{
		U:                      u,
		MaxDownloadCount:       runtime.NumCPU() * 2,
		Debug:                  true,
		Output:                 os.TempDir(),
		Cookies:                "",
		PerRequestIntervalTime: 10,
	}
}
