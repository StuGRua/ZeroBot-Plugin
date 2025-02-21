package minecraftobserver

import (
	"fmt"
	"testing"
)

func Test_formatSubStatusChange(t *testing.T) {

}

func Test_singleServerScan(t *testing.T) {
	initErr := initializeDB("data/minecraftobserver/" + dbPath)
	if initErr != nil {
		t.Fatalf("initializeDB() error = %v", initErr)
	}
	if db == nil {
		t.Fatalf("initializeDB() got = %v, want not nil", db)
	}
	t.Run("状态变更", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &ServerSubscribeSchema{
			ServerAddr:  "cn.nekoland.top",
			TargetGroup: 123456,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "",
		}
		err := db.insertServerSubscribe(newSS1)
		if err != nil {
			t.Fatalf("upsertServerStatus() error = %v", err)
		}
		ss2, err := db.getServerSubscribeByTargetGroupAndAddr("cn.nekoland.top", 123456)
		if err != nil {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() error = %v", err)
		}
		changed, msg, err := singleServerScan(ss2)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if !changed {
			t.Fatalf("singleServerScan() got = %v, want true", changed)
		}
		if len(msg) == 0 {
			t.Fatalf("singleServerScan() got = %v, want not empty", msg)
		}
		fmt.Printf("msg: %v\n", msg)
	})

	t.Run("不可达", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &ServerSubscribeSchema{
			ServerAddr:  "dx.123213213123123.net",
			TargetGroup: 123456,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "",
			PingDelay:   123,
		}
		err := db.insertServerSubscribe(newSS1)
		if err != nil {
			t.Fatalf("upsertServerStatus() error = %v", err)
		}
		ss2, err := db.getServerSubscribeByTargetGroupAndAddr("dx.123213213123123.net", 123456)
		if err != nil {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() error = %v", err)
		}
		changed, msg, err := singleServerScan(ss2)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if !changed {
			t.Fatalf("singleServerScan() got = %v, want true", changed)
		}
		if len(msg) == 0 {
			t.Fatalf("singleServerScan() got = %v, want not empty", msg)
		}
		fmt.Printf("msg: %v\n", msg)
	})

}
