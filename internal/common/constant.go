package common

type WorkerStatus int32

const (
	Idle    WorkerStatus = iota // 空闲
	Working                     // 正常执行中
	Risking                     // 风控等待
	Down                        // down机
)

// 任务状态
type TaskStatus string

const (
	TaskStatusPending TaskStatus = "Pending" //需要重新分配
	TaskStatusDoing   TaskStatus = "Doing"
	TaskStatusDone    TaskStatus = "Done"
)

func (s WorkerStatus) String() string {
	return [...]string{"Idle", "Working", "Risking", "Down"}[s]
}
