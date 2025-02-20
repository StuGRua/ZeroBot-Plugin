package minecraftobserver

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/FloatTech/floatbox/file"
	"github.com/FloatTech/gg"
	"github.com/FloatTech/zbputils/control"
	"github.com/FloatTech/zbputils/img/text"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/message"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"strings"
	"time"
)

// serverStatusDTO 服务器状态数据传输对象 From mc server response
type serverStatusDTO struct {
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

// getMinecraftServerStatus 获取Minecraft服务器状态
func getMinecraftServerStatus(addr string) (*serverStatusDTO, error) {
	resp, delay, err := bot.PingAndList(addr)
	if err != nil {
		logrus.Errorf("[mcobserver] PingAndList error: %+v", err)
		return nil, err
	}
	var s serverStatusDTO
	err = json.Unmarshal(resp, &s)
	if err != nil {
		logrus.Errorf("[drawServerStatus] Parse json response fail: %+v", err)
		return nil, err
	}
	s.Delay = delay
	return &s, nil
}

// generateServerStatusPicMsg 生成服务器状态图片消息
func (dto *serverStatusDTO) generateServerStatusPicMsg() (msg message.Segment, err error) {
	// 绘制图片
	info, err := dto.drawServerStatus()
	if err != nil {
		logrus.Errorf("[mc-ob] drawAndGenerateServerStatusNoticeMessage error: %v", err)
		return
	}
	// 发送图片，控制图片大小
	buffer := bytes.NewBuffer(make([]byte, 0, 1024*1024*4)) // 4MB
	err = jpeg.Encode(buffer, info, &jpeg.Options{Quality: 100})
	if err != nil {
		logrus.Errorf("[mc-ob] drawAndGenerateServerStatusNoticeMessage error: %v", err)
		return
	}
	msg = message.ImageBytes(buffer.Bytes())
	return
}

const (
	pingListPicTotalWidth  = 800
	pingListPicTotalHeight = 200
)

// drawServerStatus 绘制服务器状态的图片
func (dto *serverStatusDTO) drawServerStatus() (img image.Image, err error) {
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
	favicon, fErr := dto.Favicon.toImage()
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
	onlineInfo := fmt.Sprintf("在线人数：\t%d\t/\t%d", dto.Players.Online, dto.Players.Max)
	canvas.DrawString(onlineInfo, 200, 70)
	ver := fmt.Sprintf("版本：\t%s", dto.Version.Name)
	logrus.Infof("[drawServerStatus] ver: %v", ver)
	canvas.DrawString(ver, 200, 90)
	// 需要处理不可达的情况
	if dto.Delay < 0 {
		canvas.SetRGBA255(255, 0, 0, 255)
		canvas.DrawString("Ping：\t连接失败", 200, 110)
	} else {
		canvas.DrawString(fmt.Sprintf("Ping：\t%d 毫秒", dto.Delay.Milliseconds()), 200, 110)
	}
	canvas.SetColor(color.White)
	drawColoredText(canvas, dto.Description.String(), 50, 150, pingListPicTotalWidth/2)

	img = canvas.Image()

	return img, nil
}
