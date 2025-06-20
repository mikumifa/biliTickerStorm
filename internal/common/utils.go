package common

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
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
	l.SetFormatter(&customFormatter{prefix: prefix})
	log = &Logger{Logger: l, prefix: prefix}
	return log
}

// customFormatter 实现 logrus.Formatter
type customFormatter struct {
	prefix string
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 时间格式：2025-06-20T16:06:50.123
	timestamp := entry.Time.Format("2006-01-02T15:04:05.000")

	// 颜色设置，prefix 颜色青色，level 根据等级设置颜色
	prefixColored := color.New(color.FgCyan).Sprint(f.prefix)

	levelColor := color.New()
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = color.New(color.FgBlue)
	case logrus.InfoLevel:
		levelColor = color.New(color.FgGreen)
	case logrus.WarnLevel:
		levelColor = color.New(color.FgYellow)
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = color.New(color.FgRed)
	default:
		levelColor = color.New(color.FgCyan)
	}
	level := levelColor.Sprintf("%s", entry.Level.String())

	// 拼接格式：时间 | 前缀 | 级别 | 消息\n
	logLine := fmt.Sprintf("%s|%s|%s|%s\n", timestamp, prefixColored, level, entry.Message)
	return []byte(logLine), nil
}
