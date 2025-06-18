package common

import (
	"log"
	"os"
	"syscall"
)

func SeedSignalExit() {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatalf("Failed to find process: %v", err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		log.Fatalf("Failed to send SIGTERM: %v", err)
	}
}
