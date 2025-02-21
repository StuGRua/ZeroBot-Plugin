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
		Brief:            "Minecraft服务器状态查询/订阅",
		// 详细帮助
		Help: "- mc服务器状态 [IP/URI]\n" +
			"- mc服务器添加订阅 [IP/URI]\n" +
			"- mc服务器删除订阅 [IP/URI]\n" +
			"- 拉取mc服务器订阅 （仅限群聊，需要插件定时任务配合使用）" +
			"-----------------------\n" +
			"使用job插件设置定时, 例:" +
			"记录在\"@every 1m\"触发的指令\n" +
			"mc服务器订阅拉取",
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
	engine.OnRegex("^[m|M][c|C]服务器状态 (.+)$").SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		logrus.Infoln(logPrefix+"extractedPlainText: ", extractedPlainText)
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器状态 ", "")
		resp, err := getMinecraftServerStatus(addr)
		if err != nil || resp == nil {
			logrus.Errorln(logPrefix+"getMinecraftServerStatus error: ", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		status := resp.GenServerSubscribeSchema(addr, 0, 0)
		msg := status.GenerateServerStatusMsg()
		if id := ctx.SendChain(msg...); id.ID() == 0 {
			ctx.SendChain(message.Text("发送失败..."))
			return
		}
	})
	// 添加订阅
	engine.OnRegex(`^[m|M][c|C]服务器添加订阅\s*(.+)$`, zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器添加订阅 ", "")
		ss, err := getMinecraftServerStatus(addr)
		if err != nil || ss == nil {
			logrus.Errorln(logPrefix+"getMinecraftServerStatus error: ", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		// 插入数据库
		err = dbInstance.insertServerSubscribe(ss.GenServerSubscribeSchema(addr, 0, ctx.Event.GroupID))
		if err != nil {
			logrus.Errorln(logPrefix+"insertServerSubscribe error: ", err)
			ctx.SendChain(message.Text("订阅添加失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.SendChain(message.Text("订阅添加成功"))
	})
	// 删除
	engine.OnRegex(`^[m|M][c|C]服务器删除订阅\s*(.+)$`, zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器删除订阅 ", "")
		// 通过群组id和服务器地址获取服务器状态
		ss, err := dbInstance.getServerSubscribeByTargetGroupAndAddr(addr, ctx.Event.GroupID)
		if err != nil {
			logrus.Errorln(logPrefix+"getServerSubscribeByTargetGroupAndAddr error: ", err)
			ctx.SendChain(message.Text("查询订阅失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		if ss == nil {
			ctx.SendChain(message.Text("查询订阅失败...", fmt.Sprintf("错误信息: %v", "未找到订阅")))
			return
		}
		// 删除数据库
		err = dbInstance.deleteServerSubscribeByID(ss.ID)
		if err != nil {
			logrus.Errorln(logPrefix+"deleteServerStatus error: ", err)
			ctx.SendChain(message.Text("订阅删除失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		ctx.SendChain(message.Text("订阅删除成功"))
	})
	// 状态变更通知，仅限群聊使用
	engine.OnRegex(`^[m|M][c|C]服务器订阅拉取$`, zero.OnlyGroup, getDB).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		serverList, err := dbInstance.getAllServerSubscribeByTargetGroup(ctx.Event.GroupID)
		if err != nil {
			logrus.Errorln(logPrefix+"getAllServerSubscribeByTargetGroup error: ", err)
			return
		}
		changedCount := 0
		for _, oldSubStatus := range serverList {
			isChanged, changedNotifyMsg, sErr := singleServerScan(oldSubStatus)
			if sErr != nil {
				logrus.Errorln(logPrefix+"singleServerScan error: ", sErr)
				continue
			}
			if !isChanged {
				continue
			}
			changedCount++
			logrus.Infoln(logPrefix+"singleServerScan changed in ", oldSubStatus.ServerAddr, " - ", oldSubStatus.TargetGroup)
			// 发送变化信息
			if id := ctx.SendChain(changedNotifyMsg...); id.ID() == 0 {
				logrus.Errorln(logPrefix + "SendChain error id is 0")
				continue
			}
		}
		logrus.Infoln(logPrefix+"拉取mc服务器订阅 获取到: ", changedCount, "个服务器订阅状态变更")
	})
}

const (
	subStatusChangeTextNoticeTitleFormat = "Minecraft服务器状态变更通知:\n"
	// 图标变更
	subStatusChangeTextNoticeIconFormat = "图标变更:\n"
)

func formatSubStatusChange(oldStatus, newStatus *ServerSubscribeSchema) (msg message.Message) {
	msg = make(message.Message, 0)
	if oldStatus == nil || newStatus == nil {
		return
	}
	if oldStatus.Description != newStatus.Description {
		msg = append(msg, message.Text(fmt.Sprintf("描述变更: %v -> %v\n", oldStatus.Description, newStatus.Description)))
	}
	if oldStatus.Version != newStatus.Version {
		msg = append(msg, message.Text(fmt.Sprintf("版本变更: %v -> %v\n", oldStatus.Version, newStatus.Version)))
	}
	if oldStatus.FaviconMD5 != newStatus.FaviconMD5 {
		msg = append(msg, message.Text(subStatusChangeTextNoticeIconFormat))
		var faviconOldBase64, faviconNewBase64 string
		if oldStatus.FaviconRaw.checkPNG() {
			faviconOldBase64 = oldStatus.FaviconRaw.toBase64String()
			msg = append(msg, message.Text("旧图标："), message.Image(faviconOldBase64), message.Text("->"))
		} else {
			msg = append(msg, message.Text("旧图标：无->"))
		}
		if newStatus.FaviconRaw.checkPNG() {
			faviconNewBase64 = newStatus.FaviconRaw.toBase64String()
			msg = append(msg, message.Text("新图标："), message.Image(faviconNewBase64), message.Text("\n"))
		} else {
			msg = append(msg, message.Text("新图标：无\n"))
		}
	}
	// 状态由不可达变为可达，反之
	if oldStatus.PingDelay == PingDelayUnreachable && newStatus.PingDelay != PingDelayUnreachable {
		msg = append(msg, message.Text(fmt.Sprintf("Ping延迟：超时 -> %d\n", newStatus.PingDelay)))
	}
	if oldStatus.PingDelay != PingDelayUnreachable && newStatus.PingDelay == PingDelayUnreachable {
		msg = append(msg, message.Text(fmt.Sprintf("Ping延迟：%d -> 超时\n", oldStatus.PingDelay)))
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
		logrus.Warnf(logPrefix+"getMinecraftServerStatus error: %v", err)
		err = nil
		// 深拷贝，设置PingDelay为不可达
		oldSubStatus.DeepCopyTo(newSubStatus)
		newSubStatus.PingDelay = PingDelayUnreachable
	} else {
		// 没有错误则更新服务器状态
		newSubStatus = rawServerStatus.GenServerSubscribeSchema(oldSubStatus.ServerAddr, oldSubStatus.ID, oldSubStatus.TargetGroup)
	}
	if newSubStatus == nil {
		logrus.Errorln(logPrefix + "newSubStatus is nil")
		return
	}
	// 检查是否有订阅信息变化
	if oldSubStatus.IsSubscribeSpecChanged(newSubStatus) {
		logrus.Warnf(logPrefix+"server subscribe spec changed: (%+v) -> (%+v)", oldSubStatus, newSubStatus)
		changed = true
		// 更新数据库
		err = dbInstance.updateServerSubscribeStatus(newSubStatus)
		if err != nil {
			logrus.Errorln(logPrefix+"updateServerSubscribeStatus error: ", err)
			return
		}
		// 服务状态
		newStatusMsg := newSubStatus.GenerateServerStatusMsg()
		// 发送变化信息 + 服务状态信息
		notifyMsg = append(notifyMsg, formatSubStatusChange(oldSubStatus, newSubStatus)...)
		notifyMsg = append(notifyMsg, message.Text("\n当前状态:\n"))
		notifyMsg = append(notifyMsg, newStatusMsg...)
	}
	return
}
