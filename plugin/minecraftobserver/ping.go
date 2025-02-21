package minecraftobserver

import (
	"encoding/json"
	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
	"time"
)

// getMinecraftServerStatus 获取Minecraft服务器状态
func getMinecraftServerStatus(addr string) (*ServerPingAndListResp, error) {
	resp, delay, err := bot.PingAndListTimeout(addr, time.Second*5)
	if err != nil {
		logrus.Errorln(logPrefix+"PingAndList error: ", err)
		return nil, err
	}
	var s ServerPingAndListResp
	err = json.Unmarshal(resp, &s)
	if err != nil {
		logrus.Errorln(logPrefix+"Parse json response fail: ", err)
		return nil, err
	}
	s.Delay = delay
	return &s, nil
}
