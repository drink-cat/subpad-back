package model

import "gorm.io/gorm"

// Store 汇总各表的增删改查。数据库未配置时为 nil。
type Store struct {
	UserInfo   *UserInfoRepo
	TokenInfo  *TokenInfoRepo
	SubpadInfo *SubpadInfoRepo
	FeeInfo    *FeeInfoRepo
	SyncEvent  *SyncEventRepo
	SyncCursor *SyncCursorRepo
}

func NewStore(db *gorm.DB) *Store {
	if db == nil {
		return nil
	}
	return &Store{
		UserInfo:   &UserInfoRepo{db: db},
		TokenInfo:  &TokenInfoRepo{db: db},
		SubpadInfo: &SubpadInfoRepo{db: db},
		FeeInfo:    &FeeInfoRepo{db: db},
		SyncEvent:  &SyncEventRepo{db: db},
		SyncCursor: &SyncCursorRepo{db: db},
	}
}
