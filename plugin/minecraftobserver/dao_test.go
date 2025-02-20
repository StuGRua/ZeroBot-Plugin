package minecraftobserver

import (
	"testing"
)

func cleanTestData(t *testing.T) {
	err := db.sdb.Table(tableServerSubscribe).Delete(&ServerSubscribeSchema{}).Where("id > 0").Error
	if err != nil {
		t.Fatalf("cleanTestData() error = %v", err)
	}
}

func Test_DAO(t *testing.T) {
	initErr := initializeDB("data/minecraftobserver/" + dbPath)
	if initErr != nil {
		t.Fatalf("initializeDB() error = %v", initErr)
	}
	if db == nil {
		t.Fatalf("initializeDB() got = %v, want not nil", db)
	}
	t.Run("insert", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &ServerSubscribeSchema{
			ServerAddr:  "dx.zhaomc.net",
			TargetGroup: 123456,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		newSS2 := &ServerSubscribeSchema{
			ServerAddr:  "dx.zhaomc.net",
			TargetGroup: 777777,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := db.insertServerSubscribe(newSS1)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = db.insertServerSubscribe(newSS2)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		// check insert
		queryResult, err := db.getAllServerSubscribeByTargetGroup(123456)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult) != 1 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 1", len(queryResult))
		}
		if queryResult[0].TargetGroup != 123456 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 123456", queryResult[0].TargetGroup)
		}

		// check insert 2
		queryResult2, err := db.getAllServerSubscribeByTargetGroup(777777)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult2) != 1 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 1", len(queryResult2))
		}
		if queryResult2[0].TargetGroup != 777777 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 777777", queryResult2[0].TargetGroup)
		}

		// 点查
		queryResult3, err := db.getServerSubscribeByTargetGroupAndAddr("dx.zhaomc.net", 123456)
		if err != nil {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() error = %v", err)
		}
		if queryResult3.TargetGroup != 123456 {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() got = %v, want 123456", queryResult3.TargetGroup)
		}

	})
	t.Run("update", func(t *testing.T) {
		cleanTestData(t)
		newSS := &ServerSubscribeSchema{
			ServerAddr:  "dx.zhaomc.net",
			TargetGroup: 123456,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := db.insertServerSubscribe(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		// check insert
		queryResult, err := db.getAllServerSubscribeByTargetGroup(123456)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult) != 1 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 1", len(queryResult))
		}
		if queryResult[0].TargetGroup != 123456 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 123456", queryResult[0].TargetGroup)
		}
		queryResult[0].Description = "更新测试"
		err = db.updateServerSubscribeStatus(queryResult[0])
		if err != nil {
			t.Errorf("updateServerSubscribeStatus() error = %v", err)
		}
		// check update
		queryResult2, err := db.getAllServerSubscribeByTargetGroup(123456)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult2) != 1 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 1", len(queryResult2))
		}
		if queryResult2[0].Description != "更新测试" {
			t.Errorf("updateServerSubscribeStatus() got = %v, want 更新测试", queryResult2[0].Description)
		}
	})
	t.Run("delete", func(t *testing.T) {
		cleanTestData(t)
		newSS := &ServerSubscribeSchema{
			ServerAddr:  "dx.zhaomc.net",
			TargetGroup: 123456,
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := db.insertServerSubscribe(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		// check insert
		queryResult, err := db.getAllServerSubscribeByTargetGroup(123456)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult) != 1 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 1", len(queryResult))
		}
		if queryResult[0].TargetGroup != 123456 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 123456", queryResult[0].TargetGroup)
		}
		err = db.deleteServerSubscribeById(queryResult[0].ID)
		if err != nil {
			t.Errorf("deleteServerStatus() error = %v", err)
		}
		// check delete
		queryResult2, err := db.getAllServerSubscribeByTargetGroup(123456)
		if err != nil {
			t.Errorf("getAllServerSubscribeByTargetGroup() error = %v", err)
		}
		if len(queryResult2) != 0 {
			t.Errorf("getAllServerSubscribeByTargetGroup() got = %v, want 0", len(queryResult2))
		}
	})
}
