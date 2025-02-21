package minecraftobserver

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"github.com/FloatTech/imgfactory"
	"github.com/pkg/errors"
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
	// Title
	Title string `json:"title" gorm:"column:title;default:''"`
	// 纯净的服务器描述（不含修饰符）
	Description string `json:"description" gorm:"column:description;default:''"`
	// 在线玩家
	Players string `json:"players" gorm:"column:players;default:''"`
	// 版本
	Version string `json:"version" gorm:"column:version;default:''"`
	// Favicon MD5
	FaviconMD5 string `json:"favicon_md5" gorm:"column:favicon_md5;default:''"`
	// 原始数据 blob
	FaviconRaw []byte `json:"favicon_raw" gorm:"column:favicon_raw;default:null"`
	// 延迟，不可达时为-1
	PingDelay int64 `json:"ping_delay" gorm:"column:ping_delay;default:-1"`
	// 更新时间
	LastUpdate int64 `json:"last_update" gorm:"column:last_update;default:0"`
}

const (
	ColNamePingDelay  = "ping_delay"
	ColNameLastUpdate = "last_update"
)

const (
	// PingDelayUnreachable 不可达
	PingDelayUnreachable = -1
)

// isSubscribeSpecChanged 检查是否有订阅信息变化
func (ss *ServerSubscribeSchema) isSubscribeSpecChanged(new *ServerSubscribeSchema) bool {
	if ss == nil || new == nil {
		return false
	}
	// 描述变化、版本变化、Favicon变化
	if ss.Description != new.Description || ss.Version != new.Version || ss.FaviconMD5 != new.FaviconMD5 {
		return true
	}
	return false
}

// DeepCopy 深拷贝
func (ss *ServerSubscribeSchema) DeepCopy(dst *ServerSubscribeSchema) {
	dst.ID = ss.ID
	dst.ServerAddr = ss.ServerAddr
	dst.TargetGroup = ss.TargetGroup
	dst.Description = ss.Description
	dst.Players = ss.Players
	dst.Version = ss.Version
	dst.FaviconMD5 = ss.FaviconMD5
	dst.PingDelay = ss.PingDelay
	dst.LastUpdate = ss.LastUpdate
}

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

// DB Schema End
// ====================

// GenServerSubscribeSchema 将DTO转换为DB Schema
func (dto *serverPingAndListResp) GenServerSubscribeSchema(id int64, targetGroupID int64) *ServerSubscribeSchema {
	faviconMD5 := md5.Sum(helper.StringToBytes(string(dto.Favicon)))
	return &ServerSubscribeSchema{
		ID:          id,
		ServerAddr:  dto.Description.ClearString(),
		TargetGroup: targetGroupID,
		Title:       dto.Description.Text,
		Description: dto.Description.ClearString(),
		Version:     dto.Version.Name,
		FaviconMD5:  hex.EncodeToString(faviconMD5[:]),
		FaviconRaw:  []byte(dto.Favicon),
		PingDelay:   dto.Delay.Milliseconds(),
		LastUpdate:  time.Now().Unix(),
	}
}
