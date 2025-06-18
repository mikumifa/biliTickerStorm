package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"time"
	_ "time/tzdata"
)

func (w *Worker) Buy(ctx context.Context, ticketsInfo BiliTickerBuyConfig, timeStart string, interval int, pushplusToken string) error {
	Logger.WithFields(logrus.Fields{
		"detail":        ticketsInfo.Detail,
		"timeStart":     timeStart,
		"interval":      interval,
		"pushplusToken": pushplusToken,
		"Username":      ticketsInfo.Username,
	}).Info("任务信息")
	if timeStart == "" {
		timeStart = os.Getenv("TICKET_TIME_START")
		if timeStart == "" {
			Logger.Warnf("未设置环境变量 TICKET_TIME_START，不使用定时抢票")
		}
	}
	if pushplusToken == "" {
		pushplusToken = os.Getenv("PUSHPLUS_TOKEN")
		if pushplusToken == "" {
			Logger.Warnf("未设置环境变量 TICKET_TIME_START，pushplus提醒")
		}
	}
	if interval == -1 {
		intervalStr := os.Getenv("TICKET_INTERVAL")
		if intervalStr == "" {
			Logger.Warnf("未设置环境变量 TICKET_INTERVAL，将使用默认重试间隔 300")
			interval = 300
		} else {
			parsed, err := strconv.Atoi(intervalStr)
			if err != nil || parsed <= 0 {
				Logger.Warnf("TICKET_INTERVAL 格式错误（不是数字或者小于等于0），将使用默认值 300")
				interval = 300
			} else {
				interval = parsed
				Logger.Infof("从环境变量加载重试间隔: %d 秒", interval)
			}
		}
	}
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
	if timeStart != "" {
		Logger.Infof("等待开始时间 :%s", timeStart)
		loc, _ := time.LoadLocation("Asia/Shanghai")
		startTime, err := time.ParseInLocation("2006-01-02T15:04", timeStart, loc)
		if err != nil {
			return fmt.Errorf("时间格式错误: %v，正确格式应为 2006-01-02T15:04（北京时间）", err)
		}
		err = SleepUntilAccurate(startTime)
		if err != nil {
			return fmt.Errorf("等待出现问题: %w", ctx.Err())
		}
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("任务被取消: %w", ctx.Err())
		default:
		}
		Logger.Info("1）订单准备")
		prepareURL := fmt.Sprintf("https://show.bilibili.com/api/ticket/order/prepare?project_id=%d", ticketsInfo.ProjectId)
		resp, err := client.Post(prepareURL, tokenPayload)
		if err != nil {
			Logger.Errorf("读取响应失败: %v", err)
			continue
		}
		var requestResult map[string]interface{}
		if err := json.Unmarshal(resp, &requestResult); err != nil {
			Logger.Errorf("解析响应失败: %s", string(resp))
			continue
		}
		code := getIntFromMap(requestResult, "errno", "code")
		if code == -401 {
			Logger.Info("检测到验证码，调用验证码服务处理")
			err := HandleCaptcha(client, requestResult, ticketsInfo.Phone)
			if err != nil {
				Logger.Info("验证码失败")
			} else {
				Logger.Info("过验证码失败")
			}
			continue
		}
		if data, ok := requestResult["data"].(map[string]interface{}); ok {
			if token, ok := data["token"].(string); ok {
				ticketsInfo.Token = token
			}
		}
		Logger.Info("2）创建订单")
		ticketsInfo.Again = 1
		ticketsInfo.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		createURL := fmt.Sprintf("https://show.bilibili.com/api/ticket/order/createV2?project_id=%d", ticketsInfo.ProjectId)
		var errno int
		for attempt := 1; attempt <= 60; attempt++ {
			body, err := ticketsInfo.ToCreateV2RequestBody()
			if err != nil {
				Logger.Errorf("[尝试 %d/60] 创建CreateV2请求体失败: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			resp, err := client.Post(createURL, body)
			if err != nil {
				Logger.Errorf("[尝试 %d/60] 请求异常: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			var ret map[string]interface{}
			if err := json.Unmarshal(resp, &ret); err != nil {
				Logger.Errorf("[尝试 %d/60] 解析响应失败: %v", attempt, err)
				time.Sleep(time.Duration(interval) * time.Millisecond)
				continue
			}
			errno = getIntFromMap(ret, "errno", "code")
			errMsg := errnoDict[errno]
			if errMsg == "" {
				errMsg = "未知错误码"
			}
			Logger.WithFields(logrus.Fields{
				"attempt": attempt,
				"errno":   errno,
				"errMsg":  errMsg,
			}).Info("CreateV2")
			if errno == 100034 {
				if data, ok := ret["data"].(map[string]interface{}); ok {
					if payMoney, ok := data["pay_money"].(float64); ok {
						Logger.Infof("更新票价为：%.2f", payMoney/100)
						ticketsInfo.PayMoney = int(payMoney)
					}
				}
			}
			//抢票成功
			if errno == 0 {
				Logger.Info("3）抢票成功，请前往订单中心查看")
				if pushplusToken != "" {
					err := sendPushPlusMessage(pushplusToken, "抢票成功", "前往订单中心付款吧")
					if err != nil {
						return err
					}
				}
				break
			}
			if errno == 100048 || errno == 100079 {
				Logger.Info("已经下单，有尚未完成订单")
				break
			}
			if errno == 100051 {
				Logger.Info("订单准备过期，重新验证")
				break
			}

			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
		if errno == 100051 {
			Logger.Info("token过期，需要重新准备订单")
			continue
		}
		if errno == 0 {
			break
		}
	}

	return nil
}
