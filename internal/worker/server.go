package worker

import (
	"biliTickerStorm/internal/worker/pb"
	"context"
	"github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedTicketWorkerServer
	worker *Worker
	logger *logrus.Logger
}

func NewServer(worker *Worker) *Server {
	return &Server{worker: worker, logger: Logger}
}
func (s *Server) PushTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {

	err := s.worker.RunTask(ctx, req.TicketsInfo, s.worker.m.workerID)
	if err != nil {
		return &pb.TaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.TaskResponse{
		Success: true,
		Message: "Task accepted",
	}, nil
}
