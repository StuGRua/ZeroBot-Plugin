package minecraftobserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"image"
	"image/png"
	"strings"
	"time"
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

type serverStatus struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	Favicon Icon
	Delay   time.Duration
}

// Icon should be a PNG image that is Base64 encoded
// (without newlines: \n, new lines no longer work since 1.13)
// and prepended with "data:image/png;base64,".
type Icon string

func (i Icon) toImage() (icon image.Image, err error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(string(i), prefix) {
		return nil, errors.Errorf("server icon should prepended with %s", prefix)
	}
	base64png := strings.TrimPrefix(string(i), prefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
	icon, err = png.Decode(r)
	return
}
