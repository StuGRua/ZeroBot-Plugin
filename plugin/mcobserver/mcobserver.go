// Package mcobserver 通过mc服务器地址获取服务器状态信息并绘制图片发送到QQ群
package mcobserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/floatbox/web"
	"github.com/FloatTech/gg"
	ctrl "github.com/FloatTech/zbpctrl"
	"github.com/FloatTech/zbputils/control"
	zbpCtxExt "github.com/FloatTech/zbputils/ctxext"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/Tnze/go-mc/bot"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"time"
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
		resp, delay, err := bot.PingAndList(addr)
		if err != nil {
			logrus.Errorf("[mcobserver] PingAndList error: %v", err)
			ctx.SendChain(message.Text("服务器状态获取失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		info, err := makePicForPingListInfo(resp, delay)
		if err != nil {
			logrus.Errorf("[mcobserver] makePicForPingListInfo error: %v", err)
			ctx.SendChain(message.Text("绘制状态图失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*4)) // 4MB
		err = jpeg.Encode(buffer, info, &jpeg.Options{Quality: 100})
		if err != nil {
			logrus.Errorf("[mcobserver] makePicForPingListInfo error: %v", err)
			ctx.SendChain(message.Text("绘制状态图失败...", fmt.Sprintf("错误信息: %v", err)))
			return
		}
		if id := ctx.SendChain(message.ImageBytes(buffer.Bytes())); id.ID() == 0 {
			ctx.SendChain(message.Text("发送状态失败..."))
			return
		}
	})
}

const (
	pingListPicTotalWidth  = 800
	pingListPicTotalHeight = 200
)

func makePicForPingListInfo(resp []byte, delay time.Duration) (img image.Image, err error) {
	var s status
	err = json.Unmarshal(resp, &s)
	if err != nil {
		fmt.Print("[makePicForPingListInfo] Parse json response fail:", err)
		return nil, err
	}
	s.Delay = delay

	canvas := gg.NewContext(pingListPicTotalWidth, pingListPicTotalHeight)

	backgroundData, gErr := web.GetData("http://tva1.sinaimg.cn/large/0066094Sgy1hgbztp9e0pj30m805k0tm.jpg")
	if gErr != nil {
		canvas.SetColor(color.White)
		canvas.Clear()
	} else {
		background, _, dErr := image.Decode(bytes.NewReader(backgroundData))
		if dErr != nil {
			canvas.SetColor(color.White)
			canvas.Clear()
		}
		canvas.DrawImage(background, 0, 0)
	}

	// favicon
	favicon, fErr := s.Favicon.toImage()
	if fErr != nil {
		logrus.Errorf("[makePicForPingListInfo] favicon to image error: %v", fErr)
	} else {
		canvas.DrawImage(favicon, 70, 50)
	}
	fontByte, err := file.GetLazyData(text.SakuraFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	err = canvas.ParseFontFace(fontByte, 20)
	if err != nil {
		logrus.Errorf("[makePicForPingListInfo] ParseFontFace error: %v", err)
		return
	}
	canvas.SetColor(color.White)
	onlineInfo := fmt.Sprintf("在线人数：\t%d\t/\t%d", s.Players.Online, s.Players.Max)
	canvas.DrawString(onlineInfo, 200, 70)
	ver := fmt.Sprintf("版本：\t%s", s.Version.Name)
	logrus.Infof("[makePicForPingListInfo] ver: %v", ver)
	canvas.DrawString(ver, 200, 90)
	canvas.DrawString(fmt.Sprintf("Ping：\t%d 毫秒", s.Delay.Milliseconds()), 200, 110)
	canvas.SetColor(color.White)
	drawColoredText(canvas, s.Description.String(), 50, 150, pingListPicTotalWidth/2)

	img = canvas.Image()

	return img, nil
}
