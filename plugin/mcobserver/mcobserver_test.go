package mcobserver

import (
	"fmt"
	"github.com/FloatTech/imgfactory"
	"testing"
)

// dx.zhaomc.net
func Test_makePicForPingListInfo(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		ss, err := getMinecraftServerStatus("dx.zhaomc.net")
		if err != nil {
			t.Errorf("getMinecraftServerStatus() error = %v", err)
		}
		gotImg, err := drawServerStatus(ss)
		if err != nil {
			t.Errorf("drawServerStatus() error = %v", err)
		}
		if err = imgfactory.SavePNG2Path("test.png", gotImg); err != nil {
			t.Errorf("imgfactory.Save() error = %v", err)
		}
	})

}

func Test_ExampleIcon_ToImagex(t *testing.T) {
	// 示例字符串
	ansiString := "\x1b[32mHello\x1b[0m, \x1b[31mworld\x1b[0m!"
	// 解析字符串
	parsed := parseFormatText(ansiString)
	// 打印解析结果
	for _, p := range parsed {
		fmt.Printf("Text: %s, Format: %s\n", p.text, p.ctl)
	}
}
