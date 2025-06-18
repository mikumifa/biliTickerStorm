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

var log = Logger

type Manager struct {
	mu           sync.Mutex
	workerID     string
	address      string
	masterAddr   string
	status       WorkerStatus
	TaskAssigned string
	stopChan     chan struct{}
}

func (wm *Manager) GetStatus() WorkerStatus {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.status
}

func (wm *Manager) SetStatus(s WorkerStatus, taskId string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	wm.status = s
	wm.TaskAssigned = taskId
}

func NewWorkerManager(masterAddr string) *Manager {
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("worker-%s-%d", hostname, time.Now().Unix())

	return &Manager{
		workerID:   workerID,
		masterAddr: masterAddr,
		status:     Idle,
		stopChan:   make(chan struct{}),
	}
}

func (wm *Manager) RegisterToMaster() error {
	var err error
	wm.address, err = GetOutboundIPToMaster(wm.masterAddr)
	if err != nil {
		return fmt.Errorf("连接获取本地IP失败: %v", err)
	}
	wm.address += ":40051"

	conn, err := grpc.Dial(wm.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("连接主服务器失败: %v", err)
	}

	client := masterpb.NewTicketMasterClient(conn)

	req := &masterpb.WorkerInfo{
		WorkerId:     wm.workerID,
		Address:      wm.address,
		Status:       int32(wm.GetStatus()),
		TaskAssigned: wm.TaskAssigned,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reply, err := client.RegisterWorker(ctx, req)
	if err != nil {
		return fmt.Errorf("注册请求失败: %v", err)
	}

	if !reply.Success {
		return fmt.Errorf("注册失败: %s", reply.Message)
	}

	log.Printf("成功注册到主服务器: WorkerID=%s, Address=%s", wm.workerID, wm.address)
	return nil
}

func (wm *Manager) StartHeartbeat(interval time.Duration) {
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

func (wm *Manager) sendHeartbeat() error {
	conn, err := grpc.Dial(wm.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := masterpb.NewTicketMasterClient(conn)
	req := &masterpb.WorkerInfo{
		WorkerId:     wm.workerID,
		Address:      wm.address,
		Status:       int32(wm.GetStatus()),
		TaskAssigned: wm.TaskAssigned,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.RegisterWorker(ctx, req)
	if err != nil {
		log.Printf("心跳更新失败: %v", err)
	}
	return err
}
func (wm *Manager) UpdateTaskStatus(status TaskStatus, taskId string) error {
	conn, err := grpc.Dial(wm.masterAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := masterpb.NewTicketMasterClient(conn)
	req := &masterpb.TaskStatusUpdate{
		WorkerId: wm.workerID,
		Status:   string(status),
		TaskId:   taskId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.UpdateTaskStatus(ctx, req)
	if err != nil {
		log.Printf("更新Task失败: %v", err)
	}
	return err
}

// UpdateWorkerStatus 更新状态，同时通过心跳汇报给worker taskId和状态绑定，在更新状态时候同步更新taskId
func (wm *Manager) UpdateWorkerStatus(status WorkerStatus, taskId string) error {
	wm.SetStatus(status, taskId)
	return wm.sendHeartbeat()
}

func (wm *Manager) Stop() {
	close(wm.stopChan)
}

func (wm *Manager) GetWorkerID() string {
	return wm.workerID
}
