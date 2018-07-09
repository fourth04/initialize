package model

import (
	"github.com/fourth04/initialize/restfulbygin/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	_ "github.com/lib/pq"
)

var DB *gorm.DB
var err error

func init() {
	// Openning file
	DB, err = gorm.Open(config.Cfg.Dialect, config.Cfg.DBPath)
	// Display SQL queries
	DB.LogMode(config.Cfg.DBLogMode)

	// Error
	if err != nil {
		panic(err)
	}
	// Creating the table
	if !DB.HasTable(&User{}) {
		DB.CreateTable(&User{})
		// DB.Set("gorm:table_options", "ENGINE=InnoDB").CreateTable(&User{})
	}
}
