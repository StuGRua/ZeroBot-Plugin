package minecraftobserver

//const (
//	pathObserver  = "data/MinecraftObserver/"
//	fnBackground1 = "bg1_800_200.jpeg"
//)
//
//// getBackGroundData 获取背景图数据，优先使用懒加载数据，失败则使用源站地址
//// 未来可能支持多种背景图
//func getBackGroundData() (data []byte, err error) {
//	//data, err = file.GetLazyData(pathObserver+fnBackground1, control.Md5File, false)
//	//if err != nil {
//	//	logrus.Errorf("[mcobserver] getBackGroundData from lazy-data error: %v", err)
//	//}
//	// 回源
//	if len(data) == 0 {
//		data, err = web.GetData("http://tva1.sinaimg.cn/large/0066094Sgy1hgbztp9e0pj30m805k0tm.jpg")
//		if err != nil {
//			logrus.Errorf("[mcobserver] getBackGroundData from back-up addr error: %v", err)
//			return
//		}
//	}
//	return
//}
