package model

import (
	"time"

	"github.com/fourth04/initialize/restfulbygin/utils"
)

type User struct {
	ID            uint      `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`
	Username      string    `gorm:"TYPE:VARCHAR(100);UNIQUE_INDEX;NOT NULL" form:"username" json:"username"`
	Password      string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"password" json:"password"`
	Salt          string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"salt" json:"salt"`
	RoleName      string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"role_name" json:"role_name"`
	RateFormatted string    `gorm:"TYPE:VARCHAR(100);NOT NULL" form:"rate_formatted" json:"rate_formatted"`
	CreatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `gorm:"DEFAULT:CURRENT_TIMESTAMP"`
}

func (u User) Desentitize() map[string]interface{} {
	userDesensitized := utils.StructToMap(u)
	delete(userDesensitized, "Password")
	delete(userDesensitized, "Salt")
	return userDesensitized
}
