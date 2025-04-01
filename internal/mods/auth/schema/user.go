package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"size:50;not null;uniqueIndex;comment:用户名"`
	Email     string    `json:"email" gorm:"size:100;not null;uniqueIndex;comment:邮箱"`
	Password  string    `json:"-" gorm:"size:100;not null;comment:密码"`
	Avatar    string    `json:"avatar" gorm:"size:255;comment:头像URL"`
	Bio       string    `json:"bio" gorm:"type:text;comment:个人简介"`
	Role      string    `json:"role" gorm:"size:20;not null;default:user;comment:角色"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"comment:更新时间"`
}

// TableName 表名
func (a *User) TableName() string {
	return config.C.FormatTableName("user")
}

// UserResponse 用户响应结构
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar"`
	Bio       string    `json:"bio"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token LoginToken   `json:"token"`
	User  UserResponse `json:"user"`
}

type LoginToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresAt   int64  `json:"expires_at"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Username    string `json:"username" binding:"omitempty,min=3,max=20"`
	Email       string `json:"email" binding:"omitempty,email"`
	Bio         string `json:"bio"`
	Avatar      string `json:"avatar"`
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

// Captcha 验证码
type Captcha struct {
	CaptchaID string `json:"captcha_id"`
}
