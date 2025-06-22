package worker

import (
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	netUrl "net/url"
	"strings"
	"time"
)

func (bc *BiliClient) getCookieValue(name string) string {
	for _, cookie := range bc.cookies {
		if strings.EqualFold(cookie.Name, name) {
			return cookie.Value
		}
	}
	return ""
}

type BiliClient struct {
	client  *fasthttp.Client
	cookies []Cookies
	worker  *Worker
}

func NewBiliClient(cookies []Cookies, worker *Worker) *BiliClient {
	return &BiliClient{
		client:  &fasthttp.Client{ReadTimeout: 30 * time.Second},
		cookies: cookies,
		worker:  worker,
	}
}

func (bc *BiliClient) setHeaders(req *fasthttp.Request) {
	h := &req.Header
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	h.Set("Content-Type", "application/json")
	h.Set("Referer", "https://show.bilibili.com/")
	var cookieStr string
	for _, c := range bc.cookies {
		if c.Domain == ".bilibili.com" {
			cookieStr += c.Name + "=" + c.Value + "; "
		}
	}
	if cookieStr != "" {
		h.Set("Cookie", cookieStr)
	}
}

func (bc *BiliClient) Get(url string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.Header.SetMethod("GET")
	req.SetRequestURI(url)
	bc.setHeaders(req)

	if err := bc.client.Do(req, resp); err != nil {
		return nil, err
	}
	err := bc.handleHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (bc *BiliClient) Post(url string, data interface{}) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return resp.Body(), err
	}
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	req.SetBody(jsonData)
	bc.setHeaders(req)

	if err := bc.client.Do(req, resp); err != nil {
		return nil, err
	}
	err = bc.handleHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
func (bc *BiliClient) DoFormRequest(url string, data map[string]string) ([]byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)
	req.Header.SetMethod("POST")
	req.SetRequestURI(url)
	bc.setHeaders(req)
	req.Header.SetContentType("application/x-www-form-urlencoded")
	form := netUrl.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	req.SetBodyString(form.Encode())
	client := &fasthttp.Client{}
	err := client.Do(req, resp)
	if err != nil {
		return nil, err
	}
	err = bc.handleHTTPStatus(resp)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (bc *BiliClient) handleHTTPStatus(resp *fasthttp.Response) error {
	status := resp.StatusCode()
	switch status {
	case fasthttp.StatusOK:
		return nil
	case fasthttp.StatusPreconditionFailed:
		if bc.worker.cancel != nil {
			bc.worker.cancel() //取消
		}
		return fmt.Errorf("412风控")
	case fasthttp.StatusTooManyRequests:
		return fmt.Errorf("429请求过多")
	default:
		return fmt.Errorf("HTTP %d: %s", status, resp.Body())
	}
}
