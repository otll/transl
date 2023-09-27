package baidutransl

import (
	"encoding/json"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type BaiduResult struct {
	TransResult struct {
		Data []struct {
			Dst string `json:"dst"`
		} `json:"data"`
	} `json:"trans_result"`
}

func getBaiduResult(baiduResult BaiduResult) string {
	return baiduResult.TransResult.Data[0].Dst
}

func loadsResult(result []byte) BaiduResult {
	var baiduResult BaiduResult
	err := json.Unmarshal(result, &baiduResult)
	if err != nil {
		log.Fatal(err)
	}
	return baiduResult
}

func baiduTranslateTokenCookie(cookies string) (string, string) {
	client := &http.Client{}
	url := "https://fanyi.baidu.com/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	if cookies != "" {
		req.Header.Set("Cookie", cookies)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	cookiesStr := ""
	if cookies != "" {
		cookiesStr = cookies
	} else {
		for _, c := range resp.Cookies() {
			cookiesStr += c.Name + "=" + c.Value + ";"
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	tokenPattern, err := regexp.Compile(`token: '(.*?)',`)
	if err != nil {
		log.Fatal(err)
	}
	tokenStr := tokenPattern.FindStringSubmatch(string(body))
	if len(tokenStr) == 2 {
		return tokenStr[1], cookiesStr
	}

	return "", cookiesStr
}

func getTokenCookie() (string, string) {
	_, cookies := baiduTranslateTokenCookie("")
	return baiduTranslateTokenCookie(cookies)
}

func getSign(keyword string) string {

	_, filePath, _, _ := runtime.Caller(0)
	dirPath := path.Dir(filePath)
	script := openFile(filepath.Join(dirPath, "baidu.js"))

	vm := goja.New()
	_, err := vm.RunString(script)
	if err != nil {
		panic(err)
	}
	var getSign func(string, string) string
	err = vm.ExportTo(vm.Get("get_sign"), &getSign)
	if err != nil {
		panic(err)
	}
	return getSign(keyword, "320305.131321201")
}

func openFile(name string) string {
	content, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func Transl(keyword string) string {
	token, cookies := getTokenCookie()
	sign := getSign(keyword)

	body := strings.NewReader(fmt.Sprintf(`from=en&to=zh&query=%s&transtype=realtime&simple_means_flag=3&sign=%s&token=%s&domain=common&ts=%v`, keyword, sign, token, time.Now().UnixMilli()))
	req, err := http.NewRequest("POST", "https://fanyi.baidu.com/v2transapi?from=en&to=zh", body)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36")
	req.Header.Set("Acs-Token", "xxx")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Cookie", cookies)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return getBaiduResult(loadsResult(result))
}
