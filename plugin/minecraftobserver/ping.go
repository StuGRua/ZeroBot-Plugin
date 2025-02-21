package minecraftobserver

import (
	"encoding/json"
	"fmt"
	"github.com/RomiChan/syncx"
	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	// pingServerUnreachableCounter Ping服务器不可达计数器，防止bot本体网络抖动导致误报
	pingServerUnreachableCounter = syncx.Map[string, _pingServerUnreachableCounter]{}
	// 计数器阈值
	pingServerUnreachableCounterThreshold = int64(3)
	// 时间阈值
	pingServerUnreachableCounterTimeThreshold = time.Minute * 30
)

type _pingServerUnreachableCounter struct {
	count              int64
	firstUnreachableTs time.Time
}

func addPingServerUnreachableCounter(groupID int64, addr string, ts time.Time) (afterAdded int64, getTs time.Time) {
	key := fmt.Sprintf("%d-%s", groupID, addr)
	get, ok := pingServerUnreachableCounter.Load(key)
	if !ok {
		pingServerUnreachableCounter.Store(key, _pingServerUnreachableCounter{
			count:              1,
			firstUnreachableTs: ts,
		})
		return 1, ts
	}
	// 存在则更新，时间戳不变
	pingServerUnreachableCounter.Store(key, _pingServerUnreachableCounter{
		count:              get.count + 1,
		firstUnreachableTs: get.firstUnreachableTs,
	})
	return get.count + 1, get.firstUnreachableTs
}

func resetPingServerUnreachableCounter(groupID int64, addr string) {
	key := fmt.Sprintf("%d-%s", groupID, addr)
	pingServerUnreachableCounter.Delete(key)
}

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
