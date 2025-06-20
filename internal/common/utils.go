package common

import (
	formatter "github.com/DaRealFreak/colored-nested-formatter"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var (
	log  *Logger
	once sync.Once
)

type Logger struct {
	*logrus.Logger
	prefix string
}

func GetLogger(prefix string) *Logger {
	once.Do(func() {
		log = NewLogger(prefix)
	})
	return log
}

// NewLogger 返回带前缀的 Logger 实例
func NewLogger(prefix string) *Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.DebugLevel)
	l.SetFormatter(&formatter.Formatter{
		DisableColors:            false,
		ForceColors:              false,
		DisableTimestamp:         false,
		UseUppercaseLevel:        true,
		UseTimePassedAsTimestamp: false,
		TimestampFormat:          time.StampMilli,
		PadAllLogEntries:         true,
	})
	log = &Logger{Logger: l, prefix: prefix}
	return log
}

// customFormatter 实现 logrus.Formatter
type customFormatter struct {
	prefix string
}
