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
	dbPath               = "minecraft_observer.dbInstance"
)

var errDBConn = errors.New("数据库连接失败")

type db struct {
	sdb  *gorm.DB
	lock sync.RWMutex
}

// initializeDB 初始化数据库
func initializeDB(dbpath string) error {
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
	dbInstance = &db{
		sdb:  gdb,
		lock: sync.RWMutex{},
	}
	return nil
}

var (
	// dbInstance 数据库实例
	dbInstance *db
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
func (d *db) getAllServerSubscribeByTargetGroup(targetGroupID int64) ([]*ServerSubscribeSchema, error) {
	if d == nil {
		return nil, errDBConn
	}
	var ss []*ServerSubscribeSchema
	if err := d.sdb.Table(tableServerSubscribe).Where("target_group = ?", targetGroupID).Find(&ss).Error; err != nil {
		return nil, err
	}
	return ss, nil
}

// 通过群组id和服务器地址获取订阅
func (d *db) getServerSubscribeByTargetGroupAndAddr(addr string, targetGroupID int64) (*ServerSubscribeSchema, error) {
	if d == nil {
		return nil, errDBConn
	}
	var ss ServerSubscribeSchema
	if err := d.sdb.Table(tableServerSubscribe).Where("server_addr = ? and target_group = ?", addr, targetGroupID).Find(&ss).Error; err != nil {
		logrus.Errorf("[mc-ob] getServerSubscribeByTargetGroupAndAddr ERROR: %v", err)
		return nil, err
	}
	return &ss, nil
}

func (d *db) updateServerSubscribeStatus(ss *ServerSubscribeSchema) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if ss == nil {
		return errors.New("参数错误")
	}
	if ss.ID == 0 {
		return errors.New("ID不能为空")
	}
	if ss.LastUpdate == 0 {
		ss.LastUpdate = time.Now().Unix()
	}
	if err = d.sdb.Table(tableServerSubscribe).Model(ss).Update(ss).Where("id = ?", ss.ID).Error; err != nil {
		logrus.Errorf("[mc-ob] updateServerSubscribeStatus ERROR: %v", err)
		return
	}
	return
}

func (d *db) insertServerSubscribe(ss *ServerSubscribeSchema) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if ss == nil {
		return errors.New("参数错误")
	}
	if ss.LastUpdate == 0 {
		ss.LastUpdate = time.Now().Unix()
	}
	if err = d.sdb.Table(tableServerSubscribe).Create(ss).Error; err != nil {
		logrus.Errorf("[mc-ob] insertServerSubscribe ERROR: %v", err)
		return
	}
	return
}

func (d *db) deleteServerSubscribeByID(id int64) (err error) {
	if d == nil {
		return errDBConn
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	if id == 0 {
		return errors.New("ID不能为空")
	}
	if err = d.sdb.Table(tableServerSubscribe).Delete(&ServerSubscribeSchema{}).Where("id = ?", id).Error; err != nil {
		logrus.Errorf("[mc-ob] deleteServerSubscribeByID ERROR: %v", err)
		return
	}
	return
}
