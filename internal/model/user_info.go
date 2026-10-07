package model

import (
	"context"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserInfo struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Username string `gorm:"size:200;index:idx_user_info_username" json:"username"`
	Password string `gorm:"size:200" json:"-"`
	FeeAddr  string `gorm:"size:100" json:"fee_addr"`
}

func (UserInfo) TableName() string { return "user_info" }

func (u *UserInfo) BeforeSave(*gorm.DB) error {
	if u == nil || u.Password == "" || isBcryptHash(u.Password) {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	u.Password = string(hash)
	return nil
}

func (u *UserInfo) CheckPassword(plain string) error {
	if u == nil || u.Password == "" {
		return fmt.Errorf("check password: empty")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plain)); err != nil {
		return fmt.Errorf("check password: %w", err)
	}
	return nil
}

func isBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}

type UserInfoFilter struct {
	Username string
	Page
}

type UserInfoRepo struct {
	db *gorm.DB
}

func (r *UserInfoRepo) Create(ctx context.Context, row *UserInfo) error {
	if row == nil {
		return fmt.Errorf("create user_info: nil row")
	}
	if row.Username == "" || row.Password == "" {
		return fmt.Errorf("create user_info: username and password are required")
	}
	return createRow(ctx, r.db, "user_info", row)
}

func (r *UserInfoRepo) Get(ctx context.Context, id int64) (*UserInfo, error) {
	return getRow[UserInfo](ctx, r.db, "user_info", id)
}

func (r *UserInfoRepo) GetByUsername(ctx context.Context, username string) (*UserInfo, error) {
	if username == "" {
		return nil, fmt.Errorf("get user_info: username is required")
	}
	var row UserInfo
	err := r.db.WithContext(ctx).Where("username = ?", username).Order("id ASC").First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get user_info: %w", err)
	}
	return &row, nil
}

func (r *UserInfoRepo) Update(ctx context.Context, row *UserInfo) error {
	if row == nil {
		return fmt.Errorf("update user_info: nil row")
	}
	return updateRow(ctx, r.db, "user_info", row.ID, row)
}

func (r *UserInfoRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[UserInfo](ctx, r.db, "user_info", id)
}

func (r *UserInfoRepo) List(ctx context.Context, f UserInfoFilter) ([]UserInfo, error) {
	q := r.db.WithContext(ctx).Model(&UserInfo{})
	if f.Username != "" {
		q = q.Where("username = ?", f.Username)
	}
	rows := make([]UserInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list user_info: %w", err)
	}
	return rows, nil
}
