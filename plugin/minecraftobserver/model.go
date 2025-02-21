package minecraftobserver

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/FloatTech/imgfactory"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wdvxdr1123/ZeroBot/message"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"
	"image"
	"image/png"
	"strings"
	"time"
)

// ====================
// DB Schema

// ServerSubscribeSchema 服务器订阅信息
type ServerSubscribeSchema struct {
	// ID 主键
	ID int64 `json:"id" gorm:"column:id;primary_key:pk_id;auto_increment;default:0"`
	// 服务器地址
	ServerAddr string `json:"server_addr" gorm:"column:server_addr;default:'';unique_index:idx_server_addr_target_group"`
	// 信息推送群组
	TargetGroup int64 `json:"target_group" gorm:"column:target_group;default:0;unique_index:idx_server_addr_target_group;index:idx_target_group"`
	// 服务器描述
	Description string `json:"description" gorm:"column:description;default:null;type:CLOB"`
	// 在线玩家
	Players string `json:"players" gorm:"column:players;default:''"`
	// 版本
	Version string `json:"version" gorm:"column:version;default:''"`
	// FaviconMD5 Favicon MD5
	FaviconMD5 string `json:"favicon_md5" gorm:"column:favicon_md5;default:''"`
	// FaviconRaw 原始数据
	FaviconRaw Icon `json:"favicon_raw" gorm:"column:favicon_raw;default:null;type:CLOB"`
	// 延迟，不可达时为-1
	PingDelay int64 `json:"ping_delay" gorm:"column:ping_delay;default:-1"`
	// 更新时间
	LastUpdate int64 `json:"last_update" gorm:"column:last_update;default:0"`
}

const (
	// PingDelayUnreachable 不可达
	PingDelayUnreachable = -1
)

// IsSubscribeSpecChanged 检查是否有订阅信息变化
func (ss *ServerSubscribeSchema) IsSubscribeSpecChanged(newStatus *ServerSubscribeSchema) (res bool) {
	res = false
	if ss == nil || newStatus == nil {
		res = false
		return
	}
	// 描述变化、版本变化、Favicon变化
	if ss.Description != newStatus.Description || ss.Version != newStatus.Version || ss.FaviconMD5 != newStatus.FaviconMD5 {
		res = true
		return
	}
	// 状态由不可达变为可达 or 反之
	if (ss.PingDelay == PingDelayUnreachable && newStatus.PingDelay != PingDelayUnreachable) ||
		(ss.PingDelay != PingDelayUnreachable && newStatus.PingDelay == PingDelayUnreachable) {
		res = true
		return
	}
	return
}

// DeepCopyTo 深拷贝
func (ss *ServerSubscribeSchema) DeepCopyTo(dst *ServerSubscribeSchema) {
	if ss == nil || dst == nil {
		return
	}
	dst.ID = ss.ID
	dst.ServerAddr = ss.ServerAddr
	dst.TargetGroup = ss.TargetGroup
	dst.Description = ss.Description
	dst.Players = ss.Players
	dst.Version = ss.Version
	dst.FaviconMD5 = ss.FaviconMD5
	dst.FaviconRaw = ss.FaviconRaw
	dst.PingDelay = ss.PingDelay
	dst.LastUpdate = ss.LastUpdate
}

// FaviconToImage 转换为 image.Image
func (ss *ServerSubscribeSchema) FaviconToImage() (icon image.Image, err error) {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(string(ss.FaviconRaw), prefix) {
		return nil, errors.Errorf("server icon should prepended with %s", prefix)
	}
	base64png := strings.TrimPrefix(string(ss.FaviconRaw), prefix)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
	icon, err = png.Decode(r)
	return
}

// FaviconToBytes ToBytes 转换为bytes
func (ss *ServerSubscribeSchema) FaviconToBytes() (b []byte, err error) {
	i, err := ss.FaviconToImage()
	if err != nil {
		return nil, err
	}
	b, err = imgfactory.ToBytes(i)
	if err != nil {
		return nil, err
	}
	return
}

// GenerateServerStatusMsg 生成服务器状态消息
func (ss *ServerSubscribeSchema) GenerateServerStatusMsg() (msg message.Message) {
	msg = make(message.Message, 0)
	if ss == nil {
		return
	}
	msg = append(msg, message.Text(fmt.Sprintf("%s\n", ss.Description)))
	// 图标
	if ss.FaviconRaw != "" && ss.FaviconRaw.checkPNG() {
		msg = append(msg, message.Image(ss.FaviconRaw.toBase64String()))
	}
	msg = append(msg, message.Text(fmt.Sprintf("版本：%s\n", ss.Version)))
	if ss.PingDelay < 0 {
		msg = append(msg, message.Text("Ping：超时\n"))
	} else {
		msg = append(msg, message.Text(fmt.Sprintf("Ping：%d 毫秒\n", ss.PingDelay)))
		msg = append(msg, message.Text(fmt.Sprintf("\n在线人数：%s\n", ss.Players)))
	}
	return
}

// DB Schema End
// ====================

// ServerPingAndListResp 服务器状态数据传输对象 From mc server response
type ServerPingAndListResp struct {
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

//func (i Icon) toImage() (icon image.Image, err error) {
//	const prefix = "data:image/png;base64,"
//	if !strings.HasPrefix(string(i), prefix) {
//		return nil, errors.Errorf("server icon should prepended with %s", prefix)
//	}
//	base64png := strings.TrimPrefix(string(i), prefix)
//	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(base64png))
//	icon, err = png.Decode(r)
//	return
//}

// checkPNG 检查是否为PNG
func (i Icon) checkPNG() bool {
	const prefix = "data:image/png;base64,"
	return strings.HasPrefix(string(i), prefix)
}

// toBase64String 转换为base64字符串
func (i Icon) toBase64String() string {
	return "base64://" + strings.TrimPrefix(string(i), "data:image/png;base64,")
}

// GenServerSubscribeSchema 将DTO转换为DB Schema
func (dto *ServerPingAndListResp) GenServerSubscribeSchema(addr string, id int64, targetGroupID int64) *ServerSubscribeSchema {
	if dto == nil {
		return nil
	}
	faviconMD5 := md5.Sum(helper.StringToBytes(string(dto.Favicon)))
	return &ServerSubscribeSchema{
		ID:          id,
		ServerAddr:  addr,
		TargetGroup: targetGroupID,
		Description: dto.Description.ClearString(),
		Version:     dto.Version.Name,
		Players:     fmt.Sprintf("%d/%d", dto.Players.Online, dto.Players.Max),
		FaviconMD5:  hex.EncodeToString(faviconMD5[:]),
		FaviconRaw:  dto.Favicon,
		PingDelay:   dto.Delay.Milliseconds(),
		LastUpdate:  time.Now().Unix(),
	}
}

// ====================

const (
	logPrefix = "[minecraft observer] "
)
