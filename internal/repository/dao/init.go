package dao

import "gorm.io/gorm"

// InitTable 初始化表
func InitTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}
