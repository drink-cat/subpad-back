package model

import "gorm.io/gorm"

const (
	defaultPageLimit = 20
	maxPageLimit     = 200
)

// Page 是列表查询的偏移和条数。Limit 小于等于 0 时用默认条数。
type Page struct {
	Offset int `form:"offset"`
	Limit  int `form:"limit"`
}

func (p Page) Apply(db *gorm.DB) *gorm.DB {
	limit := p.Limit
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	offset := p.Offset
	if offset < 0 {
		offset = 0
	}
	return db.Offset(offset).Limit(limit)
}
