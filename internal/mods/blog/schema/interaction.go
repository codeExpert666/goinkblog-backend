package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// UserInteraction 用户与文章交互模型
type UserInteraction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;comment:用户ID"`
	ArticleID uint      `json:"article_id" gorm:"not null;comment:文章ID"`
	Type      string    `json:"type" gorm:"size:20;not null;comment:交互类型"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:创建时间"`
}

// TableName 表名
func (a *UserInteraction) TableName() string {
	return config.C.FormatTableName("user_interaction")
}

// InteractionResponse 交互响应结构
type InteractionResponse struct {
	Liked     bool `json:"liked"`     // 是否已点赞
	Favorited bool `json:"favorited"` // 是否已收藏
}
