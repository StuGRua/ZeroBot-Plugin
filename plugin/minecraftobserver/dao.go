package minecraftobserver

import (
	"errors"
	fcext "github.com/FloatTech/floatbox/ctxext"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"os"
	"sync"
	"time"
)

const (
	tableServerSubscribe = "server_subscribe"
	dbPath               = "minecraft_observer.db"
)

type mcDB struct {
	sdb  *gorm.DB
	lock sync.RWMutex
}

// initializeDB 初始化数据库
func initializeDB(dbpath string) error {
	d := &mcDB{}
	d.lock.Lock()
	defer d.lock.Unlock()
	if _, err := os.Stat(dbpath); err != nil || os.IsNotExist(err) {
		// 生成文件
		f, err := os.Create(dbpath)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	gdb, err := gorm.Open("sqlite3", dbpath)
	if err != nil {
		logrus.Errorf("mc-ob] initializeDB ERROR: %v", err)
		return err
	}
	gdb.Table(tableServerSubscribe).AutoMigrate(&ServerSubscribeSchema{})
	d.sdb = gdb
	db = d
	return nil
}

var (
	db *mcDB
	// 开启并检查数据库链接
	getDB = fcext.DoOnceOnSuccess(func(ctx *zero.Ctx) bool {
		var err error
		err = initializeDB(engine.DataFolder() + dbPath)
		if err != nil {
			ctx.SendChain(message.Text("[mc-ob] ERROR: ", err))
			return false
		}
		return true
	})
)

// getAllServerSubscribeByTargetGroup 根据群组ID获取所有订阅的服务器
func (sdb *mcDB) getAllServerSubscribeByTargetGroup(targetGroupID int64) ([]*ServerSubscribeSchema, error) {
	if db == nil {
		return nil, errors.New("数据库连接失败")
	}
	var ss []*ServerSubscribeSchema
	if err := db.sdb.Table(tableServerSubscribe).Where("target_group = ?", targetGroupID).Find(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil
}

// 通过群组id和服务器地址获取订阅
func (sdb *mcDB) getServerSubscribeByTargetGroupAndAddr(addr string, targetGroupID int64) (*ServerSubscribeSchema, error) {
	if db == nil {
		logrus.Errorf("[mc-ob] getServerSubscribeByTargetGroupAndAddr ERROR: %v", "数据库连接失败")
		return nil, errors.New("数据库连接失败")
	}
	var ss ServerSubscribeSchema
	if err := db.sdb.Table(tableServerSubscribe).Where("server_addr = ? and target_group = ?", addr, targetGroupID).Find(&ss).Error; err != nil {
		logrus.Errorf("[mc-ob] getServerSubscribeByTargetGroupAndAddr ERROR: %v", err)
		return nil, err
	}
	return &ss, nil
}

func (sdb *mcDB) updateServerSubscribeStatus(ss *ServerSubscribeSchema) (err error) {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()
	if db == nil {
		return errors.New("数据库连接失败")
	}
	if ss == nil {
		return errors.New("参数错误")
	}
	if ss.ID == 0 {
		return errors.New("ID不能为空")
	}
	if ss.LastUpdate == 0 {
		ss.LastUpdate = time.Now().Unix()
	}
	if err = db.sdb.Table(tableServerSubscribe).Model(ss).Update(ss).Where("id = ?", ss.ID).Error; err != nil {
		logrus.Errorf("[mc-ob] updateServerSubscribeStatus ERROR: %v", err)
		return
	}
	return
}

//func (sdb *mcDB) setServerSubscribeStatusToUnreachable(id int64) (err error) {
//	sdb.lock.Lock()
//	defer sdb.lock.Unlock()
//	if db == nil {
//		return errors.New("数据库连接失败")
//	}
//	if err = db.sdb.Table(tableServerSubscribe).Model(&ServerSubscribeSchema{}).
//		Updates(map[string]interface{}{
//			ColNamePingDelay:  PingDelayUnreachable,
//			ColNameLastUpdate: time.Now().Unix()}).Where("id = ?", id).Error; err != nil {
//		logrus.Errorf("[mc-ob] updateServerSubscribeStatus ERROR: %v", err)
//		return
//	}
//	return
//}

func (sdb *mcDB) insertServerSubscribe(ss *ServerSubscribeSchema) (err error) {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()
	if db == nil {
		return errors.New("数据库连接失败")
	}
	if ss == nil {
		return errors.New("参数错误")
	}
	if ss.LastUpdate == 0 {
		ss.LastUpdate = time.Now().Unix()
	}
	if err = db.sdb.Table(tableServerSubscribe).Create(ss).Error; err != nil {
		logrus.Errorf("[mc-ob] insertServerSubscribe ERROR: %v", err)
		return
	}
	return
}

func (sdb *mcDB) deleteServerSubscribeById(id int64) (err error) {
	sdb.lock.Lock()
	defer sdb.lock.Unlock()
	if db == nil {
		return errors.New("数据库连接失败")
	}
	if id == 0 {
		return errors.New("ID不能为空")
	}
	if err = db.sdb.Table(tableServerSubscribe).Delete(&ServerSubscribeSchema{}).Where("id = ?", id).Error; err != nil {
		logrus.Errorf("[mc-ob] deleteServerSubscribeById ERROR: %v", err)
		return
	}
	return
}
