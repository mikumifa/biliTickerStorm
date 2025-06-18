package master

import (
	"biliTickerStorm/internal/common"
	"time"
)

type TaskInfo struct {
	ID                  string
	Status              common.TaskStatus
	AssignedTo          string // Worker ID
	TaskName            string // Ticket config file name
	TickerConfigContent string // Ticket config file content
	CreatedAt           time.Time
	UpdatedAt           time.Time
	RetryCount          int
}
