package main

import (
	"biliTickerStorm/internal/master"
	"biliTickerStorm/internal/master/pb"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs"
	}

	lis, err := net.Listen("tcp", ":40052")
	if err != nil {
		log.Fatalf("listening failed: %v", err)
	}
	masterServer := master.NewServer()
	if err := masterServer.LoadTasksFromDir(configPath); err != nil {
		log.Fatalf("Read configs failed: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterTicketMasterServer(s, masterServer)
	log.Println("BiliTickerStorm Master started successfully，listening at 40052")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Start failed: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Closing...")
	s.GracefulStop()
	log.Println("Closed")
}
