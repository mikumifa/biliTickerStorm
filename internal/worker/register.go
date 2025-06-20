package worker

import (
	. "biliTickerStorm/internal/common"
	masterpb "biliTickerStorm/internal/master/pb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"sync"
	"time"
)

type Register struct {
	mu           sync.Mutex
	workerID     string
	address      string
	masterAddr   string
	ws           WorkerStatus
	ts           TaskStatus
	TaskAssigned string
	stopChan     chan struct{}
}

func (wm *Register) GetStatus() WorkerStatus {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.ws
}

func (wm *Register) SetStatus(ws WorkerStatus, ts TaskStatus, taskId string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.ws = ws
	wm.ts = ts
	wm.TaskAssigned = taskId
}

func NewWorkerManager(masterAddr string) *Register {
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("worker-%s-%d", hostname, time.Now().Unix())

	return &Register{
		workerID:   workerID,
		masterAddr: masterAddr,
		ws:         Idle,
		stopChan:   make(chan struct{}),
	}
}

func (wm *Register) RegisterToMaster() error {
	var err error
	wm.address, err = GetOutboundIPToMaster(wm.masterAddr)
	if err != nil {
		return fmt.Errorf("连接获取本地IP失败: %v", err)
	}
	wm.address += ":40051"

	err = wm.sendHeartbeat()
	if err != nil {
		log.Errorf("注册到主服务器失败:%v", err)
	}
	log.Printf("成功注册到主服务器: WorkerID=%s, Address=%s", wm.workerID, wm.address)
	return nil
}

func (wm *Register) StartHeartbeat(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := wm.sendHeartbeat(); err != nil {
				log.Printf("心跳发送失败: %v", err)
			}
		case <-wm.stopChan:
			log.Println("停止心跳")
			return
		}
	}
}

func (wm *Register) sendHeartbeat() error {
	conn, err := grpc.Dial(wm.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := masterpb.NewTicketMasterClient(conn)
	req := &masterpb.WorkerInfo{
		WorkerId:     wm.workerID,
		Address:      wm.address,
		WorkStatus:   int32(wm.GetStatus()),
		TaskStatus:   string(wm.ts),
		TaskAssigned: wm.TaskAssigned,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.RegisterWorker(ctx, req)
	if err != nil {
		log.Errorf("%v", err)
	}
	return err
}
func (wm *Register) CancelTask(s WorkerStatus) error {
	conn, err := grpc.Dial(wm.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := masterpb.NewTicketMasterClient(conn)
	req := &masterpb.CancelTaskInfo{
		WorkerId:     wm.workerID,
		CancelTaskId: wm.TaskAssigned,
		WorkStatus:   int32(s),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.CancelTask(ctx, req)
	if err != nil {
		log.Errorf("%v", err)
	}
	return err
}

// UpdateWorkerStatusAndTaskStatus 更新 ws和ts，同时触发task的updateTime
func (wm *Register) UpdateWorkerStatusAndTaskStatus(ws WorkerStatus, ts TaskStatus, taskId string) error {
	wm.SetStatus(ws, ts, taskId)
	return wm.sendHeartbeat()
}

func (wm *Register) Stop() {
	close(wm.stopChan)
}

func (wm *Register) GetWorkerID() string {
	return wm.workerID
}
