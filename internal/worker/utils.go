package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/beevik/ntp"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var Logger = logrus.New()

func getIntFromMap(m map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if intVal, ok := val.(int); ok {
				return intVal
			}
			if floatVal, ok := val.(float64); ok {
				return int(floatVal)
			}
		}
	}
	return 0
}

func ReadFileAsString(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 发送PushPlus消息
func sendPushPlusMessage(token, title, content string) error {
	data := map[string]string{
		"token":   token,
		"title":   title,
		"content": content,
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post("http://www.pushplus.plus/send", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
func GetNestedString(m map[string]interface{}, keys ...string) (string, bool) {
	var val interface{} = m
	for _, key := range keys {
		m2, ok := val.(map[string]interface{})
		if !ok {
			return "", false
		}
		val = m2[key]
	}
	s, ok := val.(string)
	return s, ok
}

func GetAccurateTime() (time.Time, error) {
	resp, err := ntp.Query("ntp.aliyun.com")
	if err != nil {
		return time.Now(), err
	}
	accurate := time.Now().Add(resp.ClockOffset)
	return accurate, nil
}
func SleepUntilAccurate(target time.Time) error {
	for {
		now, err := GetAccurateTime()
		if err != nil {
			now = time.Now()
		}
		if now.After(target) || now.Equal(target) {
			return nil
		}
		delta := target.Sub(now)
		if delta > time.Second {
			time.Sleep(delta - time.Second)
		} else {
			runtime.Gosched()
		}
	}
}
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
func GetOutboundIPToMaster(masterAddr string) (string, error) {
	if strings.HasPrefix(masterAddr, "http://") {
		masterAddr = strings.TrimPrefix(masterAddr, "http://")
	} else if strings.HasPrefix(masterAddr, "https://") {
		masterAddr = strings.TrimPrefix(masterAddr, "https://")
	}
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return "", fmt.Errorf("dial master failed: %w", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	localIP := localAddr.IP.String()
	if strings.HasPrefix(localIP, "http://") {
		localIP = strings.TrimPrefix(localIP, "http://")
	} else if strings.HasPrefix(localIP, "https://") {
		localIP = strings.TrimPrefix(localIP, "https://")
	}
	return localIP, nil
}
