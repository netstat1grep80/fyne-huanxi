package main

import "C"
import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/andybalholm/brotli"
	"github.com/flopp/go-findfont"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var logger *log.Logger

func contentDecoding(res *http.Response) (bodyReader io.Reader, err error) {
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		bodyReader, err = gzip.NewReader(res.Body)
	case "deflate":
		bodyReader = flate.NewReader(res.Body)
	case "br":
		bodyReader = brotli.NewReader(res.Body)
	default:
		bodyReader = res.Body
	}
	return
}

func HttpRequest(url string, method string, jsonData []byte) ([]byte, error) {

	logPrintln(fmt.Sprintf("HttpRequest:%s", url))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		logPrintln(fmt.Sprintf("%s GET error : %s ", url, err.Error()))
		return nil, err
	}

	COOKIE := map[string]string{
		"Accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Encoding":    "gzip, deflate, br",
		"Accept-Language":    "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7",
		"Cache-Control":      "no-cache",
		"Connection":         "keep-alive",
		"Cookie":             "h5=1; FFNoSNoP=1; hxut=cd8eb833cfc3a05ec01bb50943bab9f60be3f466e4ea13065cc7ceadd961a9e981ab6dd57b682041b01afd1fc115c8ff457b7a6c52d93d9d72cb827e284c8ba7e0616ee6f7534be0cf7a5c2; uid=343776140; isvip=1; sisvip=0; uinfo=%7B%22username%22%3A%22%E6%AC%A2%E5%96%9C_5286%22%2C%22avatar%22%3A%22https%3A%2F%2Fpic8.huanxi.com%2F8a9eb00f7c2b2660017c965bd8f90a86.png%22%2C%22vip_normal%22%3A%7B%22start_time%22%3A1697183302118%2C%22level%22%3A1%2C%22contract_status%22%3A0%2C%22end_time%22%3A1728719302118%7D%2C%22vip_super%22%3A%7B%22start_time%22%3A0%2C%22level%22%3A0%2C%22contract_status%22%3A0%2C%22end_time%22%3A0%7D%2C%22vip_level%22%3A1%2C%22status%22%3A3%7D; tfstk=e08Mxq9K-hS1IkUW9dQs-awkuimd5R_fFKUAHZBqY9WQBRU9gS42Hpev6Ik6nEXegZh_XVaFLs5PgiW9ks8DdNSXXtBOnZAXEXH-y4d61q_reY3RRMAFqZ7gGCSk1C_bOCllOSA_o8Ll8MHgfm7iSElG-TVRBhReAhIGUBm7urYQyGXy694qUefiFTRNKE8A4gqU4_1gl6lv8oZfb61CeI8cKitF--5KtXqJGG5CNThntoNOb61hcXc32dSNO_yh.",
		"Host":               "www.huanxi.com",
		"Pragma":             "no-cache",
		"Referer":            "https://www.huanxi.com/",
		"Sec-Fetch-Dest":     "document",
		"Sec-Fetch-Mode":     "navigate",
		"Sec-Fetch-Site":     "same-origin",
		"User-Agent":         "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
		"sec-ch-ua":          "\"Google Chrome\";v=\"119\", \"Chromium\";v=\"119\", \"Not?A_Brand\";v=\"24\"",
		"sec-ch-ua-mobile":   "?0",
		"sec-ch-ua-platform": "\"macOS\"",
	}

	for k, v := range COOKIE {
		//logger.Debug(k, ": ", v)
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	res, err := client.Do(req)

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		logPrintln(fmt.Sprintf("%s client.Do error : %s ", url, err.Error()))
		return nil, err
	}

	if body, err := contentDecoding(res); err == nil {
		if result, err := io.ReadAll(body); err == nil {
			return result, nil
		}
	}
	return nil, err
}

type Huanxi struct {
	Vid         string
	Title       string
	Vtype       string
	DeviceId    string
	EncryptType string
}

func NewHuanxi(url string) (*Huanxi, error) {
	re := regexp.MustCompile(`play_(\d+).shtml`)
	match := re.FindStringSubmatch(url)
	if match == nil || len(match) < 2 {
		return nil, errors.New("url地址错误,请重新输入")
	}

	huanxi := &Huanxi{
		Vid:         match[1],
		Vtype:       "1",
		EncryptType: "1",
		DeviceId:    "1652168067513967000",
	}

	body, err := HttpRequest(url, "GET", nil)
	if err != nil {
		logPrintln(fmt.Sprintf("%s io.ReadAll error : %s ", url, err.Error()))
		return nil, err
	}

	re = regexp.MustCompile(`title:"(.*?)"`)
	//fmt.Println(string(body))
	match = re.FindStringSubmatch(string(body))
	if match == nil || len(match) < 2 {
		return nil, errors.New("无法解析当前页面，标题未匹配到")
	}
	huanxi.Title = match[1]
	return huanxi, nil
}

func (h *Huanxi) GetM3u8() (string, error) {
	m3u8 := ""
	url := fmt.Sprintf("https://www.huanxi.com/apis/hxtv/play/authen/v1?from=m&tabbar_id=1001&xt=0&platform=1&deviceId=%s&version=6.11&vid=%s&vtype=%s&encryptType=%s", h.DeviceId, h.Vid, h.Vtype, h.EncryptType)
	body, err := HttpRequest(url, "GET", nil)
	if err != nil {
		logPrintln(fmt.Sprintf("%s io.ReadAll error : %s ", url, err.Error()))
		return m3u8, err
	}

	re := regexp.MustCompile(`"cdn_url":"(.*?)"`)
	match := re.FindStringSubmatch(string(body))
	if match == nil || len(match) < 2 {
		return m3u8, errors.New("未匹配到m3u8文件")
	}
	m3u8 = match[1]
	return m3u8, nil
}

func GetCurrentDirectory() string {
	//返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Fatal(err)
	}

	return strings.Replace(dir, "\\", string(filepath.Separator), -1)
}

func executeFFmpeg(m3u8Url, startTime, duration, title string, logOutput *widget.Entry) {
	// 构建FFmpeg命令

	ffmpegCommand := []string{
		"ffmpeg",
		"-i", m3u8Url,
		"-bsf:a", "aac_adtstoasc",
		"-ss", startTime,
		"-t", duration,
		"-c", "copy",
		"-y",
		title + ".mp4",
	}

	btn.Disable()
	btn.SetText("请等待")

	logOutput.Append(logPrintln(strings.Join(ffmpegCommand, " ")))

	if runtime.GOOS == "windows" {
		//windows下，必须只能调用path下的ffmpeg路径，不能直接引用文件地址 。
		path := os.Getenv("PATH")
		path = path + ";" + GetCurrentDirectory() + string(filepath.Separator) + "lib"
		os.Setenv("PATH", path)
	}
	cmd := exec.Command(ffmpegCommand[0], ffmpegCommand[1:]...)

	var stderr bytes.Buffer
	//ffmpeg控制台输出使用stderr输出的，很奇怪
	cmd.Stderr = &stderr

	done := make(chan struct{})
	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				logOutput.Append(stderr.String())

			}
		}
	}()

	//cmd.run在整个命令执行完成之前会阻塞通道，所以上面的那个go才会一直执行下去 。
	err := cmd.Run()
	if err != nil {
		logger.Fatalf("failed to call Run(): %v", err)
	}
	btn.SetText("开始采集")
	btn.Enable()

	//done被关闭后，上面的go func也就停止执行了。
	close(done)
	logOutput.SetText("【视频采集成功】" + GetCurrentDirectory() + string(filepath.Separator) + title + ".mp4\n\n")

	return
}

func ReadCookie() (string, error) {
	fileName := GetCurrentDirectory() + string(filepath.Separator) + "cookie.txt"

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Println("无法读取文件:", err)
		return "", err
	}

	return string(data), nil
}

var btn *widget.Button

func main() {

	//设置中文字体
	fontPaths := findfont.List()
	path := ""
	for _, path = range fontPaths {
		if strings.Contains(path, "Arial Unicode.ttf") ||
			strings.Contains(path, "msyh.ttf") ||
			strings.Contains(path, "simhei.ttf") ||
			strings.Contains(path, "simsun.ttc") ||
			strings.Contains(path, "simkai.ttf") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}

	logFileName := fmt.Sprintf("%s_%s.log", "huanxi", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join("logs", logFileName)

	// 创建 logs 文件夹
	if err := os.Mkdir("logs", os.ModePerm); err != nil && !os.IsExist(err) {
		logger.Fatal(err)
		panic(any(err))
	}

	logFile, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		logger.Fatal(err)
		panic(any(err))
	}
	logger = log.New(logFile, "", log.LstdFlags)

	cookieValue, err := ReadCookie()
	if err != nil {
		//logger.Fatal(err)
		cookieValue = "错误：当前目录下未找到cookie.txt . \n" + err.Error()
	}
	app := app.New()

	w := app.NewWindow("欢喜传媒片头采集器")
	w.Resize(fyne.Size{Width: 800})
	w.CenterOnScreen()

	urlText := widget.NewEntry()
	urlText.SetPlaceHolder("播放页地址：https://www.huanxi.com/play_51973.shtml?from=m")
	urlText.MinSize()

	cookieText := widget.NewMultiLineEntry()
	cookieText.Wrapping = fyne.TextWrapBreak
	cookieText.SetMinRowsVisible(10)
	cookieText.SetText(cookieValue)
	// 创建一个多行文本输入框用于日志输出
	logOutput := widget.NewMultiLineEntry()
	//logOutput.Disable()
	logOutput.Wrapping = fyne.TextWrapBreak
	logOutput.SetMinRowsVisible(20)
	logOutput.SetText("说明:\n\t1.文本框中粘贴欢喜的播放页地址\n\t2.lib下的ffmpeg.exe不要删除\n\t3.中间出现窗体崩溃的情况查看logs下日志")
	btn = widget.NewButton("开始采集", func() {
		url := urlText.Text
		if url != "" {
			logOutput.SetText("")
			logOutput.Append(logPrintln("输入的地址:" + url))

			if huanxi, err := NewHuanxi(url); err == nil {
				if m3u8, err := huanxi.GetM3u8(); err == nil {
					logOutput.Append("获取到m3u8文件地址：" + logPrintln(m3u8))
					go executeFFmpeg(m3u8, "00:00:10", "00:06:00", huanxi.Title, logOutput)
				} else {
					logOutput.Append(logPrintln(err.Error()))
				}
			} else {
				logOutput.Append(logPrintln(err.Error()))
			}

		}
	})
	split := container.NewHSplit(urlText, btn)
	split.Offset = 0.9 //比例

	content := container.New(
		layout.NewVBoxLayout(),
		split,
		cookieText,
		logOutput,
	)
	w.SetContent(content)
	w.ShowAndRun()
}

func logPrintln(msg string) string {
	logger.Println(msg)
	return fmt.Sprintf("%s\n", msg)
}
