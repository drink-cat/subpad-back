package model

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func OpenMySQL(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, nil
	}
	db, err := gorm.Open(mysql.Open(withTimeParse(dsn)), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	return db, nil
}

// withTimeParse 让驱动把 datetime 扫成 time.Time。未设置时按 []byte 返回，查询会失败。
func withTimeParse(dsn string) string {
	lower := strings.ToLower(dsn)
	if !strings.Contains(lower, "parsetime=") {
		if strings.Contains(dsn, "?") {
			dsn += "&parseTime=true"
		} else {
			dsn += "?parseTime=true"
		}
		lower = strings.ToLower(dsn)
	}
	if !strings.Contains(lower, "loc=") {
		dsn += "&loc=Local"
	}
	return dsn
}
