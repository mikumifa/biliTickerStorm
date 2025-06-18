package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleCaptcha(client *BiliClient, requestResult map[string]interface{}, phone string) error {
	data, ok := requestResult["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无法获取data字段")
	}
	gaData, ok := data["ga_data"].(map[string]interface{})
	csrf := client.getCookieValue("bili_jct")
	if !ok {
		return fmt.Errorf("无法获取ga_data字段")
	}
	riskParams, ok := gaData["riskParams"]
	resp, err := client.Post("https://api.bilibili.com/x/gaia-vgate/v1/register", riskParams)
	if err != nil {
		return fmt.Errorf("验证码注册请求失败: %v", err)
	}
	var registerResult map[string]interface{}
	if err := json.Unmarshal(resp, &registerResult); err != nil {
		return fmt.Errorf("解析验证码注册响应失败: %v", err)
	}
	registerData, ok := registerResult["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("验证码注册响应格式错误")
	}
	token, ok := registerData["token"].(string)
	if !ok {
		return fmt.Errorf("无法获取验证码token")
	}
	captchaType, ok := registerData["type"].(string)
	if !ok {
		return fmt.Errorf("无法获取验证码类型")
	}
	var validateData map[string]interface{}
	switch captchaType {
	case "geetest":
		gt, ok := GetNestedString(registerData, "geetest", "gt")
		if !ok {
			return fmt.Errorf("无法获取gt参数")
		}
		challenge, ok := GetNestedString(registerData, "geetest", "challenge")
		if !ok {
			return fmt.Errorf("无法获取challenge参数")
		}
		validate, seccode, err := HandleGeetest(gt, challenge)
		if err != nil {
			return fmt.Errorf("极验验证码处理失败: %v", err)
		}
		requestBody := map[string]string{
			"challenge": challenge,
			"token":     token,
			"seccode":   seccode,
			"csrf":      csrf,
			"validate":  validate,
		}
		resp, err := client.DoFormRequest("https://api.bilibili.com/x/gaia-vgate/v1/validate", requestBody)
		if err != nil {
			return fmt.Errorf("极验验证请求失败: %v", err)
		}
		if err := json.Unmarshal(resp, &validateData); err != nil {
			return fmt.Errorf("解析极验验证响应失败: %v", err)
		}
	case "phone":
		if phone == "" {
			return fmt.Errorf("需要手机号码进行验证")
		}
		requestBody := map[string]interface{}{
			"code":  phone,
			"csrf":  csrf,
			"token": token,
		}
		resp, err := client.Post("https://api.bilibili.com/x/gaia-vgate/v1/validate", requestBody)
		if err != nil {
			return fmt.Errorf("手机验证请求失败: %v", err)
		}
		if err := json.Unmarshal(resp, &validateData); err != nil {
			return fmt.Errorf("解析手机验证响应失败: %v", err)
		}

	default:
		return fmt.Errorf("这是一个程序无法应对的验证码类型: %s", captchaType)
	}

	code := getIntFromMap(validateData, "errno", "code")
	if code == 0 {
		return nil
	} else {
		return fmt.Errorf("验证码失败: %v", validateData)
	}
}

func HandleGeetest(gt string, challenge string) (validate string, seccode string, err error) {
	gt_url := fmt.Sprintf("%s/validate/geetest", Cfg.GTBaseURL)
	requestBody := map[string]interface{}{
		"type":      "geetest",
		"gt":        gt,
		"challenge": challenge,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("编码请求体失败: %v", err)
	}
	resp, err := http.Post(gt_url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("请求验证服务失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("验证码服务返回状态码 %d", resp.StatusCode)
	}
	var response struct {
		Validate string `json:"validate"`
		Seccode  string `json:"seccode"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("解析响应失败: %v", err)
	}
	return response.Validate, response.Seccode, nil
}
