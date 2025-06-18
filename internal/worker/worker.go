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
	m      *Manager
	cancel context.CancelFunc
	mu     sync.Mutex // 保证并发安全地访问 cancel
}

func NewWorker(m *Manager) *Worker {
	return &Worker{
		m: m,
	}
}

func (w *Worker) RunTask(ctx context.Context, info, taskId string) error {
	workerId := w.m.workerID
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
		fmt.Println("解析 BiliTickerBuy Config 出错:", err)
		return fmt.Errorf("解析配置失败: %w", err)
	}

	go func() {
		err := w.m.UpdateWorkerStatus(Working, workerId) //set and send heartbeat
		if err != nil {
			Logger.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warning("设置状态 Working 失败")
		}
		err = w.m.UpdateTaskStatus(TaskStatusDoing, taskId)
		if err != nil {
			Logger.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warning("设置状态 TaskStatusDoing 失败")
		}
		defer func() {
			w.mu.Lock()
			w.cancel = nil
			err := w.m.UpdateWorkerStatus(Idle, taskId)
			if err != nil {
				Logger.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warning("设置状态 Idle 失败")
			}
			if err != nil {
				Logger.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warning("设置状态 TaskStatusDone 失败")
			}
			w.mu.Unlock()
		}() //执行完成
		err = w.Buy(cancelCtx, config, "", -1, "")
		if err != nil {
			Logger.WithFields(logrus.Fields{"username": config.Username, "detail": config.Detail}).Warning("抢票失败")
		}

	}()

	return nil
}
func (w *Worker) UpdateWorkerStatus(status WorkerStatus, taskId string) error {
	return w.m.UpdateWorkerStatus(status, taskId) //set and send heartbeat
}
