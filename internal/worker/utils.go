package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
)

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

func GetOutboundIPToMaster(masterAddr string) (string, error) {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		ip, err := GetLocalIP()
		if err != nil {
			return "", fmt.Errorf("dial master failed: %w", err)
		}
		return ip, nil
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.TCPAddr)
	localIP := localAddr.IP.String()
	return localIP, nil
}
func GetLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		// 过滤掉 loopback 和未启用的网卡
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no suitable IP found")
}
