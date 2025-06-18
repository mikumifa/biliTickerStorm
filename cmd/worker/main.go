package main

import (
	"biliTickerStorm/internal/common"
	workerpb "biliTickerStorm/internal/worker/pb"

	"biliTickerStorm/internal/worker"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

var log = common.Logger

func main() {
	register := worker.NewWorkerManager(worker.Cfg.MasterServerAddr) // 主服务器地址
	lis, err := net.Listen("tcp", ":40051")
	if err != nil {
		log.Fatalf("listening failed: %v", err)
	}
	workerServer := worker.NewServer(worker.NewWorker(register))
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
		if err := register.RegisterToMaster(); err != nil {
			log.Errorf("注册尝试 %d/%d 失败: %v", i+1, maxRetries, err)
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * 2 * time.Second) // 指数退避
			}
		} else {
			break
		}
	}
	go register.StartHeartbeat(3 * time.Second)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Closing...")
	err = register.CancelTask(common.Down)
	if err != nil {
		log.Errorf("%v", err)
		return
	}
	register.Stop()
	s.GracefulStop()
	log.Println("Closed")
}
