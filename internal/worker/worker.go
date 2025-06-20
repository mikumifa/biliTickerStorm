package worker

import (
	. "biliTickerStorm/internal/common"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

type Worker struct {
	m      *Register
	cancel context.CancelFunc
	mu     sync.Mutex // 保证并发安全地访问 cancel
}

func NewWorker(m *Register) *Worker {
	return &Worker{
		m: m,
	}
}

func (w *Worker) RunTask(ctx context.Context, info, taskId string) error {
	w.mu.Lock()
	if w.cancel != nil {
		w.mu.Unlock()
		return fmt.Errorf("已有任务正在执行")
	}
	cancelCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.mu.Unlock()

	var config BiliTickerBuyConfig
	if err := json.Unmarshal([]byte(info), &config); err != nil {
		log.Printf("[ConfigError] BiliTickerBuy: %v", err)
		return fmt.Errorf("解析配置失败: %w", err)
	}
	go func() {
		err := w.m.UpdateWorkerStatusAndTaskStatus(Working, TaskStatusDoing, taskId) //set and send heartbeat
		if err != nil {
			log.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warningf("设置状态 Working,TaskStatusDoing 失败: %v", err)
		}
		defer func() {
			w.mu.Lock()
			w.cancel = nil
			err := w.m.UpdateWorkerStatusAndTaskStatus(Idle, TaskStatusDone, taskId)
			if err != nil {
				log.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warningf("设置状态 Idle,TaskStatusDone失败: %v", err)
			}
			w.mu.Unlock()
		}() //执行完成
		err = w.Buy(cancelCtx, config, Cfg.TimeStart, Cfg.Interval, Cfg.PushplusToken)
		if err != nil {
			log.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warningf("抢票失败: %v", err)
		}

	}()

	return nil
}
