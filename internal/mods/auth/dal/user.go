package dal

import (
	"context"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/auth/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// GetUserDB 获取用户数据库实例
func GetUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDB(ctx, defDB).Model(&schema.User{})
}

// UserRepository 用户数据访问层
type UserRepository struct {
	DB *gorm.DB
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *schema.User) error {
	// 获取数据库实例
	db := GetUserDB(ctx, r.DB)
	// 检查用户名是否已存在
	var count int64
	if err := db.Where("username = ?", user.Username).Count(&count).Error; err != nil {
		return errors.WithStack(err)
	}
	if count > 0 {
		return errors.Conflict("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := r.DB.Model(&schema.User{}).Where("email = ?", user.Email).Count(&count).Error; err != nil {
		return errors.WithStack(err)
	}
	if count > 0 {
		return errors.Conflict("邮箱已注册")
	}

	result := db.Create(user)
	return errors.WithStack(result.Error)
}

// GetByID 通过ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id uint) (*schema.User, error) {
	db := GetUserDB(ctx, r.DB)
	var user schema.User
	err := db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

// GetByUsername 通过用户名获取用户
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*schema.User, error) {
	db := GetUserDB(ctx, r.DB)
	var user schema.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

// GetByEmail 通过邮箱获取用户
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*schema.User, error) {
	db := GetUserDB(ctx, r.DB)
	var user schema.User
	err := db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("用户不存在")
		}
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *schema.User) error {
	db := GetUserDB(ctx, r.DB)
	result := db.Where("id = ?", user.ID).Select("*").Omit("created_at").Updates(user)
	return errors.WithStack(result.Error)
}

// CheckPassword 检查密码是否正确
func (r *UserRepository) CheckPassword(ctx context.Context, user *schema.User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

// UpdatePassword 更新密码
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uint, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	db := GetUserDB(ctx, r.DB)
	result := db.Where("id = ?", userID).Update("password", string(hashedPassword))
	return errors.WithStack(result.Error)
}

// GetAll 获取所有用户
func (r *UserRepository) GetAll(ctx context.Context, page, pageSize int) ([]schema.User, int64, error) {
	db := GetUserDB(ctx, r.DB)
	var users []schema.User
	var total int64

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errors.WithStack(err)
	}

	offset := (page - 1) * pageSize
	result := db.Offset(offset).Limit(pageSize).Find(&users)
	return users, total, errors.WithStack(result.Error)
}
