package mcobserver

import (
	"encoding/json"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
)

// getMinecraftServerStatus 获取Minecraft服务器状态
func getMinecraftServerStatus(addr string) (*serverStatus, error) {
	resp, delay, err := bot.PingAndList(addr)
	if err != nil {
		logrus.Errorf("[mcobserver] PingAndList error: %v", err)
		return nil, err
	}
	var s serverStatus
	err = json.Unmarshal(resp, &s)
	if err != nil {
		fmt.Print("[drawServerStatus] Parse json response fail:", err)
		return nil, err
	}
	s.Delay = delay
	return &s, nil
}
