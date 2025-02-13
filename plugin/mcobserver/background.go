package mcobserver

import (
	"github.com/FloatTech/floatbox/web"
	"github.com/sirupsen/logrus"
)

//var pathObserver = "data/mcobserver/"

// getBackGroundData 获取背景图数据，优先使用懒加载数据，失败则使用源站地址
// 未来可能支持多种背景图
func getBackGroundData() (data []byte, err error) {
	//backgroundData, err := file.GetLazyData("background_1.jpg", control.Md5File, false)
	//if err != nil {
	//	logrus.Errorf("[mcobserver] getBackGroundData from lazy-data error: %v", err)
	//}
	if len(data) == 0 {
		data, err = web.GetData("http://tva1.sinaimg.cn/large/0066094Sgy1hgbztp9e0pj30m805k0tm.jpg")
		if err != nil {
			logrus.Errorf("[mcobserver] getBackGroundData from back-up addr error: %v", err)
			return
		}
	}
	return
}
