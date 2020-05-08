package database

import (
	"github.com/cnbattle/douyin/internal/database/model"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
)

var (
	Local        *gorm.DB
	localDialect = "sqlite3"
	localArgs    = "./database.db"
)

func init() {
	var err error
	Local, err = gorm.Open(localDialect, localArgs)
	if err != nil {
		log.Panic(err)
	}
	Local.LogMode(false)
	Local.DB().SetMaxOpenConns(10)
	Local.DB().SetMaxIdleConns(20)
	Local.AutoMigrate(&model.Video{})
}
