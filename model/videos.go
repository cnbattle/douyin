package model

import "github.com/jinzhu/gorm"

type Video struct {
	gorm.Model
	AwemeId   string `gorm:"unique;not null"`
	Nickname  string `gorm:"varchar(64)"`
	Avatar    string `gorm:"varchar(64)"`
	Desc      string `gorm:"varchar(255)"`
	CoverPath string `gorm:"varchar(255)"`
	VideoPath string `gorm:"varchar(255)"`
	Status    int    `gorm:"default(0);type:tinyint(1)"`
}
