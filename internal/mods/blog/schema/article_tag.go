package schema

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
)

// ArticleTag 文章与标签的关联表
type ArticleTag struct {
	ArticleID uint      `json:"article_id" gorm:"primaryKey;comment:文章ID"`
	TagID     uint      `json:"tag_id" gorm:"primaryKey;comment:标签ID"`
	CreatedAt time.Time `json:"created_at" gorm:"comment:创建时间"`
}

// TableName 表名
func (a *ArticleTag) TableName() string {
	return config.C.FormatTableName("article_tag")
}
