package worker

import (
	. "biliTickerStorm/internal/common"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
	_ "time/tzdata"
)

var log = GetLogger("worker")

func (w *Worker) Buy(ctx context.Context, ticketsInfo BiliTickerBuyConfig, timeStart *time.Time, interval int, pushplusToken string) error {
	log.WithFields(logrus.Fields{
		"detail":        ticketsInfo.Detail,
		"timeStart":     timeStart,
		"interval":      interval,
		"pushplusToken": pushplusToken,
		"Username":      ticketsInfo.Username,
	}).Info("接受到抢票任务")
	client := NewBiliClient(ticketsInfo.Cookies, w)
	tokenPayload := map[string]interface{}{
		"count":      ticketsInfo.Count,
		"screen_id":  ticketsInfo.ScreenId,
		"order_type": 1,
		"project_id": ticketsInfo.ProjectId,
		"sku_id":     ticketsInfo.SkuId,
		"token":      "",
		"newRisk":    true,
	}
	if timeStart != nil {
		log.Infof("开始时间 :%s", timeStart.String())
		err := SleepUntilAccurate(*timeStart)
		if err != nil {
			return err
		}
	}
	for {
		select {
		case <-ctx.Done():
			err := w.m.CancelTask(Risking)
			if err != nil {
				return err
			}
			return fmt.Errorf("任务被取消: %w", ctx.Err())
		default:
		}
		log.Info("1）订单准备")
		prepareURL := fmt.Sprintf("https://show.bilibili.com/api/ticket/order/prepare?project_id=%d", ticketsInfo.ProjectId)
		resp, err := client.Post(prepareURL, tokenPayload)
		if err != nil {
			log.Errorf("读取响应失败: %v", err)
			continue
		}
		var requestResult map[string]interface{}
		if err := json.Unmarshal(resp, &requestResult); err != nil {
			log.Errorf("解析响应失败: %s", string(resp))
			continue
		}
		code := getIntFromMap(requestResult, "errno", "code")
		if code == -401 {
			log.Info("检测到验证码，调用验证码服务处理")
			err := HandleCaptcha(client, requestResult, ticketsInfo.Phone)
			if err != nil {
				log.Info("验证码失败")
			} else {
				log.Info("过验证码失败")
			}
			continue
		}
		if data, ok := requestResult["data"].(map[string]interface{}); ok {
			if token, ok := data["token"].(string); ok {
				ticketsInfo.Token = token
			}
		}
		log.Info("2）创建订单")
		ticketsInfo.Again = 1
		ticketsInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		createURL := fmt.Sprintf("https://show.bilibili.com/api/ticket/order/createV2?project_id=%d", ticketsInfo.ProjectId)
		var errno int
		for attempt := 1; attempt <= 60; attempt++ {
			body, err := ticketsInfo.ToCreateV2RequestBody()
			if err != nil {
				log.Errorf("[尝试 %d/60] 创建CreateV2请求体失败: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			resp, err := client.Post(createURL, body)
			if err != nil {
				log.Errorf("[尝试 %d/60] 请求异常: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			var ret map[string]interface{}
			if err := json.Unmarshal(resp, &ret); err != nil {
				log.Errorf("[尝试 %d/60] 解析响应失败: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			errno = getIntFromMap(ret, "errno", "code")
			errMsg := errnoDict[errno]
			if errMsg == "" {
				errMsg = "未知错误码"
			}
			log.Infof("[Create] attempt=%d errno=%d msg=%s", attempt, errno, errMsg)
			if errno == 100034 {
				if data, ok := ret["data"].(map[string]interface{}); ok {
					if payMoney, ok := data["pay_money"].(float64); ok {
						log.Infof("更新票价为：%.2f", payMoney/100)
						ticketsInfo.PayMoney = int(payMoney)
					}
				}
			}
			//抢票成功
			if errno == 0 {
				log.Info("3）抢票成功，请前往订单中心查看")
				if pushplusToken != "" {
					err := sendPushPlusMessage(pushplusToken, "抢票成功", "前往订单中心付款吧")
					if err != nil {
						return err
					}
				}
				break
			}
			if errno == 100048 || errno == 100079 {
				log.Info("已经下单，有尚未完成订单")
				break
			}
			if errno == 100051 {
				log.Info("订单准备过期，重新验证")
				break
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
		if errno == 100051 {
			log.Info("token过期，需要重新准备订单")
			continue
		}
		if errno == 0 {
			break
		}
		log.Info("0）重新下单")

	}

	return nil
}
