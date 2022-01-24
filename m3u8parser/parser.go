package m3u8parser

import (
	"bytes"
	"fmt"
	"github.com/grafov/m3u8"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"m3u8download/options"
	"m3u8download/utils"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Parser struct {
	*options.Options
	PartsContainer chan wrapPart
}

func New(opt *options.Options) *Parser {
	return &Parser{
		Options:        opt,
		PartsContainer: make(chan wrapPart),
	}
}

func parseMediaList(u string) *m3u8.MediaPlaylist {
	mediaList := utils.Get(u)
	playlist, listType, err := m3u8.DecodeFrom(bytes.NewReader(mediaList.Body()), true)
	if err != nil {
		utils.Logger.Fatal("解析m3u8出错")
	}
	isMedia := m3u8.MEDIA == listType
	isMaster := m3u8.MASTER == listType
	if isMaster {
		masterPlaylist := playlist.(*m3u8.MasterPlaylist)
		maxBW := uint32(0)
		maxBWURI := ""
		for _, variant := range masterPlaylist.Variants {
			if variant.Bandwidth > maxBW {
				maxBW = variant.Bandwidth
				maxBWURI = variant.URI
			}
		}
		return parseMediaList(utils.ConcatDomain(u, maxBWURI))
	}

	if isMedia {
		mediaPlaylist := playlist.(*m3u8.MediaPlaylist)
		return mediaPlaylist
	}
	return nil
}

type wrapPart struct {
	Key    *m3u8.Key
	URI    string
	Sort   uint
	KeyRes string
}

func (p *Parser) Parse() {
	mediaPlaylist := parseMediaList(p.U)
	var parts []*wrapPart
	var nowKey *m3u8.Key
	// 解析m3u8连接
	for i := uint(0); i < mediaPlaylist.Count(); i++ {
		segment := mediaPlaylist.Segments[i]
		if segment.Key != nil {
			if strings.ToLower(segment.Key.Method) == "none" {
				nowKey = nil
			} else {
				nowKey = segment.Key
			}
		}

		partUri := utils.ConcatDomain(p.U, segment.URI)
		elems := &wrapPart{
			Sort:   i,
			Key:    nowKey,
			URI:    partUri,
			KeyRes: "",
		}
		parts = append(parts, elems)
		utils.Logger.Debug("解析到的part", zap.Uint("sort", elems.Sort), zap.String("URI", elems.URI))
	}
	utils.Logger.Info("总共解析的part数:", zap.Uint("count", mediaPlaylist.Count()))
	var preKey *wrapPart
	for _, part := range parts {
		part := part
		if preKey != nil && preKey.Key == part.Key {
			part.KeyRes = preKey.KeyRes
		} else {
			if part.Key != nil {
				part.KeyRes = utils.Get(utils.ConcatDomain(p.U, part.Key.URI)).String()
			}
		}
		preKey = part
	}

	go func() {
		for _, part := range parts {
			p.PartsContainer <- *part
		}
		close(p.PartsContainer)
	}()

	basePath := p.Output
	// parse, err := url.Parse(p.U)
	// if err != nil {
	// 	utils.Logger.Fatal("解析URL出错")
	// }
	outputDirName, err := ioutil.TempDir(basePath, "m3u8d-temp")
	if err != nil {
		utils.Logger.Fatal("创建临时目录出错")
	}

	// uri := parse.RequestURI()
	// urlP1 := strings.Split(uri, "#")
	// filename := strings.Split(urlP1[0], "?")[0]
	// outputDirName := filepath.Join(basePath, filename)
	// err = os.Mkdir(outputDirName, os.ModePerm)
	// if os.IsExist(err) {
	// 	utils.Logger.Fatal("文件夹已存在", zap.String("filepath", outputDirName))
	// }

	// 下载ts文件
	var wg sync.WaitGroup
	for i := 0; i < p.MaxDownloadCount; i++ {
		wg.Add(1)
		go func() {
			time.Sleep(time.Duration(rand.Intn(p.PerRequestIntervalTime)))
			for {
				part, ok := <-p.PartsContainer
				if ok {
					time.Sleep(time.Duration(p.PerRequestIntervalTime))
					fpath := filepath.Join(outputDirName, strconv.Itoa(int(part.Sort)))
					fpathExt := fmt.Sprintf("%s.ts", fpath)
					response := utils.Get(part.URI)
					parse, err := url.Parse(part.URI)
					if err != nil {
						utils.Logger.Error("提供URL错误", zap.String("url", parse.String()))
					}
					tsFile, err := os.Create(fpathExt)
					if err != nil {
						utils.Logger.Error("创建文件失败", zap.String("filepath", fpathExt))
					}
					body := response.Body()
					if part.KeyRes != "" {
						body, err = utils.AES128Decrypt(body, []byte(part.KeyRes), []byte(part.Key.IV))
						if err != nil {
							utils.Logger.Error("解析加密文件出错")
						}
					}
					_, err = tsFile.Write(body)
					if err != nil {
						utils.Logger.Error("写入文件流出错", zap.String("errfilepath", fpath))
					}
					utils.Logger.Debug("正常执行:", zap.Uint("sort", part.Sort), zap.String("url", part.URI), zap.Any("Key", part.Key), zap.String("key", part.KeyRes))
				} else {
					break
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	utils.Logger.Info("开始合并ts文件")
	dirEntries, err := os.ReadDir(outputDirName)
	if err != nil {
		utils.Logger.Fatal("读取临时文件夹失败", zap.Error(err))
	}
	tsPartCount := len(dirEntries)
	file, err := ioutil.TempFile(basePath, "m3u8d-v")
	if err != nil {
		utils.Logger.Fatal("创建输出文件失败", zap.Error(err))
	}

	for i := 0; i < tsPartCount; i++ {
		open, err := os.Open(filepath.Join(outputDirName, strconv.Itoa(i)) + ".ts")
		if err != nil {
			utils.Logger.Fatal("打开文件ts失败", zap.Error(err))
		}
		io.Copy(file, open)
		utils.Logger.Debug("ts文件合并完成", zap.Int("file", i))
	}
	defer func() {
		file.Close()
		os.RemoveAll(outputDirName)
	}()
	filename := fmt.Sprintf("%s.mp4", file.Name())
	os.Rename(file.Name(), filename)
	utils.Logger.Info("存储目录", zap.String("out", outputDirName))
	utils.Logger.Info("存储文件", zap.String("filename", filename))
}
