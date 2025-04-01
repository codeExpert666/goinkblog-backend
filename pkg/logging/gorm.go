// 定义了一个将日志写入数据库对应表的钩子执行器
package logging

import (
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

type Logger struct {
	ID        string    `gorm:"size:20;primaryKey;" json:"id"`  // 日志ID
	Level     string    `gorm:"size:20;index;" json:"level"`    // 日志级别
	TraceID   string    `gorm:"size:64;index;" json:"trace_id"` // 链路ID
	UserID    uint      `gorm:"index;" json:"user_id"`          // 用户ID
	Tag       string    `gorm:"size:32;index;" json:"tag"`      // 日志标签
	Message   string    `gorm:"size:1024;" json:"message"`      // 日志消息
	Stack     string    `gorm:"type:text;" json:"stack"`        // 错误堆栈
	Data      string    `gorm:"type:text;" json:"data"`         // 日志数据
	CreatedAt time.Time `gorm:"index;" json:"created_at"`       // 创建时间
}

func NewGormHook(db *gorm.DB) *GormHook {
	err := db.AutoMigrate(new(Logger))
	if err != nil {
		panic(err)
	}

	return &GormHook{
		db: db,
	}
}

// Gorm 日志钩子
type GormHook struct {
	db *gorm.DB
}

func (h *GormHook) Exec(extra map[string]string, b []byte) error {
	msg := &Logger{
		ID: xid.New().String(),
	}
	data := make(map[string]interface{})
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}

	if v, ok := data["ts"]; ok {
		msg.CreatedAt = time.UnixMilli(int64(v.(float64)))
		delete(data, "ts")
	}
	if v, ok := data["msg"]; ok {
		msg.Message = v.(string)
		delete(data, "msg")
	}
	if v, ok := data["tag"]; ok {
		msg.Tag = v.(string)
		delete(data, "tag")
	}
	if v, ok := data["trace_id"]; ok {
		msg.TraceID = v.(string)
		delete(data, "trace_id")
	}
	if v, ok := data["user_id"]; ok {
		if floatVal, ok := v.(float64); ok {
			msg.UserID = uint(floatVal)
		}
		delete(data, "user_id")
	}
	if v, ok := data["level"]; ok {
		msg.Level = v.(string)
		delete(data, "level")
	}
	if v, ok := data["stack"]; ok {
		msg.Stack = v.(string)
		delete(data, "stack")
	}
	delete(data, "caller")

	for k, v := range extra {
		data[k] = v
	}

	if len(data) > 0 {
		buf, _ := json.Marshal(data)
		msg.Data = string(buf)
	}

	return h.db.Create(msg).Error
}

func (h *GormHook) Close() error {
	db, err := h.db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
