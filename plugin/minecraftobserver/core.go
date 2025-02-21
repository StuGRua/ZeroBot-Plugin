// Package minecraftobserver 通过mc服务器地址获取服务器状态信息并绘制图片发送到QQ群
package minecraftobserver

import (
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"strings"
)

var (
	// 注册插件
	engine = control.Register("minecraftobserver", &ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "MC服务器状态查询/订阅",
		// 详细帮助
		Help: "- mc服务器状态 [IP/URI]\n" +
			"- 拉取mc服务器订阅 （仅限群聊，需要插件定时任务配合使用）" +
			"-----------------------\n" +
			"使用job插件设置定时, 例:" +
			"记录在\"@every 1m\"触发的指令\n" +
			"拉取mc服务器订阅",
		// 插件数据存储路径
		PrivateDataFolder: "minecraftobserver",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("minecraft observer已启动..."))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("minecraft observer已关闭..."))
		},
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

func init() {
	// 状态查询
	engine.OnRegex("^mc服务器状态 (.+)$").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		logrus.Infof("[mc-ob] extractedPlainText: (%v)", extractedPlainText)
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器状态 ", "")
		resp, err := getMinecraftServerStatus(addr)
		if err != nil || resp == nil {
			logrus.Errorf("[mc-ob] getMinecraftServerStatus error: %v", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		status := resp.GenServerSubscribeSchema(addr, 0, 0)
		msg := status.generateServerStatusMsg()
		if id := ctx.SendChain(msg...); id.ID() == 0 {
			ctx.SendChain(message.Text("发送失败..."))
			return
		}
	})
	// 添加订阅
	engine.OnRegex(`^mc服务器添加订阅\s*(.+)$`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器添加订阅 ", "")
		ss, err := getMinecraftServerStatus(addr)
		if err != nil || ss == nil {
			logrus.Errorf("[mc-ob] getMinecraftServerStatus error: %v", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		// 插入数据库
		err = db.insertServerSubscribe(ss.GenServerSubscribeSchema(addr, 0, ctx.Event.GroupID))
		if err != nil {
			logrus.Errorf("[mc-ob] insertServerSubscribe error: %v", err)
			ctx.SendChain(message.Text("订阅添加失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.SendChain(message.Text("订阅添加成功"))
	})
	// 删除
	engine.OnRegex(`^mc服务器删除订阅\s*(.+)$`, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器删除订阅 ", "")
		// 通过群组id和服务器地址获取服务器状态
		ss, err := db.getServerSubscribeByTargetGroupAndAddr(addr, ctx.Event.GroupID)
		// 删除数据库
		err = db.deleteServerSubscribeById(ss.ID)
		if err != nil {
			logrus.Errorf("[mc-ob] deleteServerStatus error: %v", err)
			ctx.SendChain(message.Text("订阅删除失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.SendChain(message.Text("订阅删除成功"))
	})
	// 状态变更通知，仅限群聊使用
	engine.OnFullMatch("拉取mc服务器订阅", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		if db == nil {
			if getDB(ctx) != true {
				logrus.Errorf("[mc-ob] getDB error: %v", "db is nil")
				return
			}
		}
		serverList, err := db.getAllServerSubscribeByTargetGroup(ctx.Event.GroupID)
		if err != nil {
			logrus.Errorf("[mc-ob] getAllServerSubscribeByTargetGroup error: %v", err)
			return
		}
		for _, oldSubStatus := range serverList {
			isChanged, changedNotifyMsg, sErr := singleServerScan(oldSubStatus)
			if sErr != nil {
				logrus.Errorf("[mc-ob] singleServerScan error: %v", sErr)
				continue
			}
			if !isChanged {
				continue
			}
			logrus.Infof("[mc-ob] singleServerScan dected changed in server: %v, group: %v", oldSubStatus.ServerAddr, oldSubStatus.TargetGroup)
			// 发送变化信息
			if id := ctx.SendChain(changedNotifyMsg...); id.ID() == 0 {
				logrus.Errorf("[mc-ob] SendChain error: %v", "id is 0")
				continue
			}
		}

	})
}

const (
	subStatusChangeTextNoticeTitleFormat = "Minecraft服务器状态变更通知:\n"
	// 标题变更
	subStatusChangeTextNoticeTitleChangeFormat = "标题变更: %v -> %v\n"
	// 描述变更
	subStatusChangeTextNoticeDescFormat = "描述变更: %v -> %v\n"
	// 版本变更
	subStatusChangeTextNoticeVersionFormat = "版本变更: %v -> %v\n"
	// 图标变更
	subStatusChangeTextNoticeIconFormat = "图标变更\n"
)

func formatSubStatusChange(old, new *ServerSubscribeSchema) (msg message.Message) {
	msg = make(message.Message, 0)
	if old == nil || new == nil {
		return
	}
	if old.Description != new.Description {
		msg = append(msg, message.Text(fmt.Sprintf(subStatusChangeTextNoticeDescFormat, old.Description, new.Description)))
	}
	if old.Version != new.Version {
		msg = append(msg, message.Text(fmt.Sprintf(subStatusChangeTextNoticeVersionFormat, old.Version, new.Version)))
	}
	if old.FaviconMD5 != new.FaviconMD5 {
		// 图标变更
		faviconOld, fErr := old.FaviconToBytes()
		if fErr != nil {
			logrus.Errorf("[mc-ob] faviconOld to image error: %v", fErr)
		}
		faviconNew, fErr := new.FaviconToBytes()
		if fErr != nil {
			logrus.Errorf("[mc-ob] faviconNew to image error: %v", fErr)
		}
		// image.Image 转 bytes
		msg = append(msg, message.Text(subStatusChangeTextNoticeIconFormat), message.Text("旧图标："),
			message.ImageBytes(faviconOld), message.Text("\n新图标："), message.ImageBytes(faviconNew), message.Text("\n"))
	}
	if len(msg) != 0 {
		msg = append([]message.Segment{message.Text(subStatusChangeTextNoticeTitleFormat)}, msg...)
	}
	return
}

// singleServerScan 单个服务器状态扫描
func singleServerScan(oldSubStatus *ServerSubscribeSchema) (changed bool, notifyMsg message.Message, err error) {
	notifyMsg = make(message.Message, 0)
	newSubStatus := &ServerSubscribeSchema{}
	// 获取服务器状态 & 检查是否需要更新
	rawServerStatus, err := getMinecraftServerStatus(oldSubStatus.ServerAddr)
	if err != nil {
		logrus.Warnf("[mc-ob] getMinecraftServerStatus error: %v", err)
		err = nil
		// 深拷贝，设置PingDelay为不可达
		oldSubStatus.DeepCopyTo(newSubStatus)
		newSubStatus.PingDelay = PingDelayUnreachable
	} else {
		// 检查是否需要更新
		newSubStatus = rawServerStatus.GenServerSubscribeSchema(oldSubStatus.ServerAddr, oldSubStatus.ID, oldSubStatus.TargetGroup)
	}
	if newSubStatus == nil {
		logrus.Errorf("[mc-ob] newSubStatus is nil")
		return
	}
	// 检查是否有订阅信息变化
	if oldSubStatus.isSubscribeSpecChanged(newSubStatus) {
		logrus.Warnf("[mc-ob] server subscribe spec changed: (%+v) -> (%+v)", oldSubStatus, newSubStatus)
		changed = true
		// 更新数据库
		err = db.updateServerSubscribeStatus(newSubStatus)
		if err != nil {
			logrus.Errorf("[mc-ob] updateServerSubscribeStatus error: %v", err)
			return
		}
		// 服务状态
		newStatusMsg := newSubStatus.generateServerStatusMsg()
		// 发送变化信息 + 服务状态信息
		notifyMsg = append(notifyMsg, formatSubStatusChange(oldSubStatus, newSubStatus)...)
		notifyMsg = append(notifyMsg, message.Text("\n当前状态:\n"))
		notifyMsg = append(notifyMsg, newStatusMsg...)
	}
	return
}
