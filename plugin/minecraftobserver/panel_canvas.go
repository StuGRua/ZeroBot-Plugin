package minecraftobserver

import (
	"bytes"
	"fmt"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/sirupsen/logrus"
	"image"
	"image/color"
)

const (
	pingListPicTotalWidth  = 800
	pingListPicTotalHeight = 200
)

// drawServerStatus 绘制服务器状态的图片
func drawServerStatus(s *serverStatus) (img image.Image, err error) {
	canvas := gg.NewContext(pingListPicTotalWidth, pingListPicTotalHeight)

	backgroundData, gErr := getBackGroundData()
	if gErr != nil {
		// 获取背景图失败，使用白色背景
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
		logrus.Errorf("[drawServerStatus] favicon to image error: %v", fErr)
	} else {
		canvas.DrawImage(favicon, 70, 50)
	}
	fontByte, err := file.GetLazyData(text.SakuraFontFile, control.Md5File, true)
	if err != nil {
		return
	}
	err = canvas.ParseFontFace(fontByte, 20)
	if err != nil {
		logrus.Errorf("[drawServerStatus] ParseFontFace error: %v", err)
		return
	}
	canvas.SetColor(color.White)
	onlineInfo := fmt.Sprintf("在线人数：\t%d\t/\t%d", s.Players.Online, s.Players.Max)
	canvas.DrawString(onlineInfo, 200, 70)
	ver := fmt.Sprintf("版本：\t%s", s.Version.Name)
	logrus.Infof("[drawServerStatus] ver: %v", ver)
	canvas.DrawString(ver, 200, 90)
	canvas.DrawString(fmt.Sprintf("Ping：\t%d 毫秒", s.Delay.Milliseconds()), 200, 110)
	canvas.SetColor(color.White)
	drawColoredText(canvas, s.Description.String(), 50, 150, pingListPicTotalWidth/2)

	img = canvas.Image()

	return img, nil
}
