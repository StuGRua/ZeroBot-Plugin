// Package mcobserver 通过mc服务器地址获取服务器状态信息并绘制图片发送到QQ群
package mcobserver

import (
	"bytes"
	"fmt"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"image/jpeg"
	"strings"
)

var (
	// 注册插件
	engine = control.Register("mcobserver", &ctrl.Options[*zero.Ctx]{
		// 默认不启动
		DisableOnDefault: false,
		Brief:            "mcobserver",
		// 详细帮助
		Help: "- mc服务器状态 [IP/URI]\n",
		// 插件数据存储路径
		PrivateDataFolder: "mcobserver",
		OnEnable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("mcobserver已启动..."))
		},
		OnDisable: func(ctx *zero.Ctx) {
			ctx.SendChain(message.Text("mcobserver已关闭..."))
		},
	}).ApplySingle(zbpCtxExt.DefaultSingle)
)

func init() {
	engine.OnRegex("^mc服务器状态 (.+)$", zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		// 关键词查找
		var extractedPlainText string
		extractedPlainText = ctx.ExtractPlainText()
		logrus.Infof("[mcobserver] extractedPlainText: (%v)", extractedPlainText)
		addr := strings.ReplaceAll(extractedPlainText, "mc服务器状态 ", "")
		ss, err := getMinecraftServerStatus(addr)
		if err != nil || ss == nil {
			logrus.Errorf("[mcobserver] getMinecraftServerStatus error: %v", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		info, err := drawServerStatus(ss)
		if err != nil {
			logrus.Errorf("[mcobserver] drawServerStatus error: %v", err)
			ctx.SendChain(message.Text("绘制状态图失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		// 发送图片，控制图片大小
		buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*4)) // 4MB
		err = jpeg.Encode(buffer, info, &jpeg.Options{Quality: 100})
		if err != nil {
			logrus.Errorf("[mcobserver] drawServerStatus error: %v", err)
			ctx.SendChain(message.Text("绘制状态图失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		if id := ctx.SendChain(message.ImageBytes(buffer.Bytes())); id.ID() == 0 {
			ctx.SendChain(message.Text("发送失败..."))
			return
		}
	})
}
