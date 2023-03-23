package rss_hub

import (
	"fmt"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/rss_hub/rss_pkg"
	"github.com/wdvxdr1123/ZeroBot/message"
	"time"
)

func formatRssFeedToTextMsg(view *rss_pkg.RssChannelView) (msg []string) {
	//logrus.Infof("RSS订阅姬：转化view %v", view.Contents)
	msg = make([]string, 0)
	// rssChannel信息
	msgStr := fmt.Sprintf("【%s】更新时间:%v\n", view.Channel.Title, view.Channel.UpdatedParsed.Format(time.DateTime))
	msg = append(msg, msgStr)
	// rssItem信息
	for _, item := range view.Contents {
		contentStr := fmt.Sprintf("标题：%s\n链接：%s\n更新时间：%v\n", item.Title, item.Link, item.Date.Format(time.DateTime))
		msg = append(msg, contentStr)
	}
	return
}

// fakeSenderForwardNode ...
func fakeSenderForwardNode(userId int64, msgs ...message.MessageSegment) message.MessageSegment {
	return message.CustomNode(
		"RssHub订阅姬",
		userId,
		msgs)
}
