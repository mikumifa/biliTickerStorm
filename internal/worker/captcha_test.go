package worker

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"
)

func TestHandleGeetest_ConcurrentPerformance(t *testing.T) {
	const concurrency = 10
	ticketsInfoStr, err := ReadFileAsString("data.json")
	if err != nil {
		t.Fatalf("读取文件出错: %v", err)
	}
	var config BiliTickerBuyConfig
	if err := json.Unmarshal([]byte(ticketsInfoStr), &config); err != nil {
		t.Fatalf("解析 JSON 出错: %v", err)
	}
	client := NewBiliClient(config.Cookies, nil)
	os.Setenv("GT_BASE_URL", "http://127.0.0.1:8000")
	defer os.Unsetenv("GT_BASE_URL")
	// 获取 gt/challenge

	var wg sync.WaitGroup
	var totalDuration time.Duration
	var mu sync.Mutex
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			get, _ := client.Get("https://passport.bilibili.com/x/passport-login/captcha?source=main_web")
			var ret map[string]interface{}
			if err := json.Unmarshal(get, &ret); err != nil {
				t.Fatalf("gt challenge 获取错误: %v", err)
			}
			gt, ok := GetNestedString(ret, "data", "geetest", "gt")
			if !ok {
				t.Fatal("gt 获取失败")
			}
			challenge, ok := GetNestedString(ret, "data", "geetest", "challenge")
			if !ok {
				t.Fatal("challenge 获取失败")
			}
			csrf := client.getCookieValue("bili_jct")
			start := time.Now()
			validate, seccode, err := HandleGeetest(gt, challenge)
			duration := time.Since(start)
			registerData, _ := ret["data"].(map[string]interface{})
			token, _ := registerData["token"].(string)
			requestBody := map[string]string{
				"challenge": challenge,
				"token":     token,
				"seccode":   seccode,
				"csrf":      csrf,
				"validate":  validate,
			}
			resp, err := client.DoFormRequest("https://api.bilibili.com/x/gaia-vgate/v1/validate", requestBody)
			if err != nil {
				t.Errorf("第 %d 个请求 HandleGeetest 返回错误: %v", i, err)
				return
			}
			var validateData map[string]interface{}
			if err := json.Unmarshal(resp, &validateData); err != nil {
			}
			if validate == "" || seccode == "" {
				t.Errorf("第 %d 个请求返回值为空: validate=%s, seccode=%s", i, validate, seccode)
			}
			code := getIntFromMap(validateData, "errno", "code")
			t.Logf("[%d] validate[%s]  code[%d]", i, validate, code)
			mu.Lock()
			totalDuration += duration
			mu.Unlock()
		}(i)
	}
	wg.Wait()
	avg := totalDuration / time.Duration(concurrency)
	t.Logf("在 %d 个并发请求下的平均响应时间: %v", concurrency, avg)
}
