package mcobserver

import (
	"github.com/FloatTech/gg"
	"image/color"
	"regexp"
	"strings"
)

var ansiColors = map[string]color.RGBA{
	"30":  {0, 0, 0, 255},       // 黑色
	"31":  {255, 0, 0, 255},     // 红色
	"32":  {0, 255, 0, 255},     // 绿色
	"33":  {255, 255, 0, 255},   // 黄色
	"34":  {0, 0, 255, 255},     // 蓝色
	"35":  {255, 0, 255, 255},   // 紫色
	"36":  {0, 255, 255, 255},   // 青色
	"37":  {255, 255, 255, 255}, // 白色
	"90":  {128, 128, 128, 255}, // 深灰色
	"91":  {255, 0, 0, 255},     // 亮红色
	"92":  {0, 255, 0, 255},     // 亮绿色
	"93":  {255, 255, 0, 255},   // 亮黄色
	"94":  {0, 0, 255, 255},     // 亮蓝色
	"95":  {255, 0, 255, 255},   // 亮紫色
	"96":  {0, 255, 255, 255},   // 亮青色
	"97":  {255, 255, 255, 255}, // 亮白色
	"40":  {0, 0, 0, 255},       // 背景黑色
	"41":  {255, 0, 0, 255},     // 背景红色
	"42":  {0, 255, 0, 255},     // 背景绿色
	"43":  {255, 255, 0, 255},   // 背景黄色
	"44":  {0, 0, 255, 255},     // 背景蓝色
	"45":  {255, 0, 255, 255},   // 背景紫色
	"46":  {0, 255, 255, 255},   // 背景青色
	"47":  {255, 255, 255, 255}, // 背景白色
	"100": {128, 128, 128, 255}, // 背景深灰色
	"101": {255, 0, 0, 255},     // 背景亮红色
	"102": {0, 255, 0, 255},     // 背景亮绿色
	"103": {255, 255, 0, 255},   // 背景亮黄色
	"104": {0, 0, 255, 255},     // 背景亮蓝色
	"105": {255, 0, 255, 255},   // 背景亮紫色
	"106": {0, 255, 255, 255},   // 背景亮青色
	"107": {255, 255, 255, 255}, // 背景亮白色
}

// 绘制带颜色的字符串
func drawColoredText(dc *gg.Context, text string, x, y, maxX float64) {
	// 分割字符串为多个部分
	parts := splitColoredText(text)
	tmpX := x
	tmpY := y
	// 逐个部分绘制
	for _, part := range parts {
		nx, ny := dc.MeasureString(part.Text)
		// 换行
		if tmpX+nx >= maxX {
			tmpY += ny * 1.5
			tmpX = x
		}
		//logrus.Infof("draw colored Text: %+v", part)
		cc := ansiColors[part.Color]
		if cc.R == 0 && cc.G == 0 && cc.B == 0 {
			dc.SetColor(color.White)
		} else {
			dc.SetRGB(float64(cc.R/100), float64(cc.G/100), float64(cc.B/100))
		}

		// 绘制文本
		dc.DrawStringAnchored(part.Text, tmpX, tmpY, 0.0, 0.0)
		tmpX += nx
	}
}

// 分割带颜色的字符串为多个部分
func splitColoredText(text string) []formatTextPart {
	var parts []formatTextPart
	pairs := parseFormatText(text)
	// 遍历划分后的字符串部分
	for _, pp := range pairs {
		tx := strings.ReplaceAll(pp.text, "\x1b[m", "")
		if pp.ctl != "" {
			pt := formatTextPart{Text: tx}
			switch pp.ctl {
			case "1":
				pt.Bold = true
			case "2":
				pt.Italic = true
			case "4":
				pt.UnderLined = true
			case "9":
				pt.StrikeThrough = true
			}
			if pt.Color == "" && (len(pp.ctl) == 2 || len(pp.ctl) == 3) {
				pt.Color = pp.ctl
			}
			parts = append(parts, pt)
		} else {
			parts = append(parts, formatTextPart{Text: pp.text, Color: ""})
		}
	}
	return parts
}

type pair struct {
	ctl  string
	text string
}

func parseFormatText(all string) []pair {
	var result []pair

	// Split the input string by the escape character
	sections := strings.Split(all, "\x1b")

	// Create a regular expression pattern to match the format code
	re := regexp.MustCompile(`(\d+(;\d+)*)`)

	// Iterate over each section except the first one
	for i := 1; i < len(sections); i++ {

		section := sections[i]

		// Find the end of the format code
		endIndex := strings.IndexAny(section, "m")

		// Extract the format code and text
		formatCode := section[:endIndex+1]
		text := section[endIndex+1:]

		// Use regular expression to find the specific ANSI control numbers
		matches := re.FindString(formatCode)

		// Append the pair to the result
		var mat string
		if len(matches) == 2 {
			mat = matches
		}
		result = append(result, pair{mat, text})
	}

	return result
}

// 表示带颜色的文本部分
type formatTextPart struct {
	Text          string
	Color         string
	Bold          bool
	Italic        bool
	UnderLined    bool
	StrikeThrough bool
}
