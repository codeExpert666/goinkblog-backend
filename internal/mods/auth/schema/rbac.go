package schema

import (
	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// CasbinRule 定义了 Casbin 规则在数据库中的存储结构
type CasbinRule struct {
	ID    uint   `json:"id" gorm:"primaryKey"`
	Ptype string `json:"ptype" gorm:"size:10;uniqueIndex:idx_casbin_rule;comment:策略类型"`
	V0    string `json:"v0" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:主体"`
	V1    string `json:"v1" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:对象"`
	V2    string `json:"v2" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:动作"`
	V3    string `json:"v3" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:扩展字段1"`
	V4    string `json:"v4" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:扩展字段2"`
	V5    string `json:"v5" gorm:"size:100;uniqueIndex:idx_casbin_rule;comment:扩展字段3"`
}

// TableName 表名
func (r *CasbinRule) TableName() string {
	return config.C.FormatTableName("casbin_rule")
}

// ========================创建操作======================

// CreatePolicyRequest 策略请求
type CreatePolicyRequest struct {
	Type    string `json:"type" binding:"required,oneof=p g"`
	Subject string `json:"subject" binding:"required"`
	Object  string `json:"object" binding:"required"`
	Action  string `json:"action" binding:"required_if=Type p"`
}

// RoleRequest 角色请求
type RoleRequest struct {
	Role string `json:"role" binding:"required"`
	User string `json:"user" binding:"required"`
}

// ========================查询操作=======================

// ListPolicyRequest 策略列表请求
type ListPolicyRequest struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"page_size" binding:"required,min=1,max=100"`
	Type     string `form:"type" binding:"omitempty,oneof=p g"`
	Subject  string `form:"subject"`
	Object   string `form:"object"`
	Action   string `form:"action"`
}

// PolicyItem 策略列表项
type PolicyItem struct {
	ID      uint   `json:"id"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action,omitempty"` // Type 为 g 时，该项不存在
}

// ListPolicyResponse 策略列表响应
type ListPolicyResponse struct {
	Total      int64         `json:"total"`
	TotalPages int           `json:"total_pages"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	Items      []*PolicyItem `json:"items"`
}

// EnforcerRequest 鉴权请求
type EnforcerRequest struct {
	Subject string `json:"subject" binding:"required"`
	Object  string `json:"object" binding:"required"`
	Action  string `json:"action" binding:"required"`
}

// EnforcerResponse 鉴权响应
type EnforcerResponse struct {
	Allowed bool `json:"allowed"`
}
