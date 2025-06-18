package main

import (
	"biliTickerStorm/internal/common"
	workerpb "biliTickerStorm/internal/worker/pb"
	"log"

	"biliTickerStorm/internal/worker"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	workerManager := worker.NewWorkerManager(worker.GetEnv("MASTER_SERVER_ADDR", "localhost:40052")) // 主服务器地址
	lis, err := net.Listen("tcp", ":40051")
	if err != nil {
		log.Fatalf("listening failed: %v", err)
	}
	workerServer := worker.NewServer(worker.NewWorker(workerManager))
	s := grpc.NewServer()
	workerpb.RegisterTicketWorkerServer(s, workerServer)
	go func() {
		log.Println("BiliTickerStorm Worker started successfully，listening at 40051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Start failed: %v", err)
		}
	}()
	time.Sleep(2 * time.Second)
	// 注册到主服务器
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err := workerManager.RegisterToMaster(); err != nil {
			log.Printf("注册尝试 %d/%d 失败: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second) // 指数退避
			}
		} else {
			log.Printf("注册成功! WorkerID: %s", workerManager.GetWorkerID())
			break
		}
	}
	go workerManager.StartHeartbeat(3 * time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Closing...")
	err = workerManager.UpdateWorkerStatus(common.Down, "")
	if err != nil {
		log.Printf("更新状态Down 失败: %s", workerManager.GetWorkerID())
		return
	}
	workerManager.Stop()
	s.GracefulStop()
	log.Println("Closed")
}
