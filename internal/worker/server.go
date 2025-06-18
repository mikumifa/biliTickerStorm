package worker

import (
	"biliTickerStorm/internal/worker/pb"
	"context"
	"fmt"
)

type Server struct {
	pb.UnimplementedTicketWorkerServer
	worker *Worker
}

func NewServer(worker *Worker) *Server {
	return &Server{worker: worker}
}
func (s *Server) PushTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {

	err := s.worker.RunTask(ctx, req.TicketsInfo, req.TaskId)
	if err != nil {
		return &pb.TaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &pb.TaskResponse{
		Success: true,
		Message: fmt.Sprintf("Task <%s> is running", req.TaskId),
	}, nil
}
