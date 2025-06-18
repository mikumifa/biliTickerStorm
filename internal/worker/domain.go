package worker

import (
	"encoding/json"
	"fmt"
)

type BiliTickerBuyConfig struct {
	Username    string      `json:"username"`
	Detail      string      `json:"detail"`
	Count       int         `json:"count"`
	ScreenId    int         `json:"screen_id"`
	ProjectId   int         `json:"project_id"`
	SkuId       int         `json:"sku_id"`
	OrderType   int         `json:"order_type"`
	PayMoney    int         `json:"pay_money"`
	BuyerInfo   []BuyerInfo `json:"buyer_info"`
	Buyer       string      `json:"buyer"`
	Tel         string      `json:"tel"`
	DeliverInfo DeliverInfo `json:"deliver_info"`
	Cookies     []Cookies   `json:"cookies"`
	Phone       string      `json:"phone"`
	Token       string      `json:"token"`
	Again       int         `json:"again"`
	Timestamp   int64       `json:"timestamp"`
}
type Cookies struct {
	Name     string  `json:"name"`
	Value    string  `json:"value"`
	Domain   string  `json:"domain"`
	Path     string  `json:"path"`
	Expires  float64 `json:"expires"`
	HttpOnly bool    `json:"httpOnly"`
	Secure   bool    `json:"secure"`
	SameSite string  `json:"sameSite"`
}
type DeliverInfo struct {
	Name   string `json:"name"`
	Tel    string `json:"tel"`
	AddrId int    `json:"addr_id"`
	Addr   string `json:"addr"`
}
type BuyerInfo struct {
	Id             int    `json:"id"`
	Uid            int    `json:"uid"`
	AccountChannel string `json:"account_channel"`
	PersonalId     string `json:"personal_id"`
	Name           string `json:"name"`
	IdCardFront    string `json:"id_card_front"`
	IdCardBack     string `json:"id_card_back"`
	IsDefault      int    `json:"is_default"`
	Tel            string `json:"tel"`
	ErrorCode      string `json:"error_code"`
	IdType         int    `json:"id_type"`
	VerifyStatus   int    `json:"verify_status"`
	AccountId      int    `json:"accountId"`
}

var errnoDict = map[int]string{
	0:      "成功",
	3:      "抢票CD中",
	100009: "库存不足,暂无余票",
	100001: "前方拥堵",
	100041: "对未发售的票进行抢票",
	100003: "验证码过期",
	100016: "项目不可售",
	100039: "活动收摊啦,下次要快点哦",
	100048: "已经下单，有尚未完成订单",
	100017: "票种不可售",
	100051: "订单准备过期，重新验证",
	100034: "票价错误",
}

type CreateV2RequestBody struct {
	Count       int    `json:"count"`
	ScreenId    int    `json:"screen_id"`
	ProjectId   int    `json:"project_id"`
	SkuId       int    `json:"sku_id"`
	OrderType   int    `json:"order_type"`
	PayMoney    int    `json:"pay_money"`
	BuyerInfo   string `json:"buyer_info"`
	Buyer       string `json:"buyer"`
	Tel         string `json:"tel"`
	DeliverInfo string `json:"deliver_info"`
	Again       int    `json:"again"`
	Token       string `json:"token"`
	Timestamp   int64  `json:"timestamp"`
}

func (cfg *BiliTickerBuyConfig) ToCreateV2RequestBody() (*CreateV2RequestBody, error) {
	buyerInfoStr, err := json.Marshal(cfg.BuyerInfo)
	if err != nil {
		return nil, fmt.Errorf("marshal buyer_info: %w", err)
	}
	deliverInfoStr, err := json.Marshal(cfg.DeliverInfo)
	if err != nil {
		return nil, fmt.Errorf("marshal deliver_info: %w", err)
	}
	return &CreateV2RequestBody{
		Count:       cfg.Count,
		ScreenId:    cfg.ScreenId,
		ProjectId:   cfg.ProjectId,
		SkuId:       cfg.SkuId,
		OrderType:   cfg.OrderType,
		PayMoney:    cfg.PayMoney,
		BuyerInfo:   string(buyerInfoStr),
		Buyer:       cfg.Buyer,
		Tel:         cfg.Tel,
		DeliverInfo: string(deliverInfoStr),
		Again:       cfg.Again,
		Timestamp:   cfg.Timestamp,
	}, nil
}
