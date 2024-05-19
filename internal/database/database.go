package database

import (
	"github.com/cnbattle/douyin/internal/database/model"
	// "gorm.io/driver/sqlite"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var (
	Local        *gorm.DB
	localDialect = "sqlite3"
	localArgs    = "./database.db"
)

func init() {
	var err error
	Local, err = gorm.Open(sqlite.Open(localArgs), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}
	Local.AutoMigrate(&model.Video{})
}
