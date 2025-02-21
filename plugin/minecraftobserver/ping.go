package minecraftobserver

import (
	"encoding/json"
	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
	"time"
)

// getMinecraftServerStatus 获取Minecraft服务器状态
func getMinecraftServerStatus(addr string) (*serverPingAndListResp, error) {
	resp, delay, err := bot.PingAndListTimeout(addr, time.Second*5)
	if err != nil {
		logrus.Errorf("[mcobserver] PingAndList error: %+v", err)
		return nil, err
	}
	logrus.Infof("[mcobserver] PingAndList response: %v", string(resp))
	var s serverPingAndListResp
	err = json.Unmarshal(resp, &s)
	if err != nil {
		logrus.Errorf("[drawServerStatus] Parse json response fail: %+v", err)
		return nil, err
	}
	s.Delay = delay
	return &s, nil
}
