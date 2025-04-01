package util

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Trans 定义了数据库事务的结构体
type Trans struct {
	DB *gorm.DB
}

// TransFunc 事务执行函数
type TransFunc func(context.Context) error

// Exec 执行数据库事务
func (a *Trans) Exec(ctx context.Context, fn TransFunc) error {
	// 检查上下文中是否已存在事务
	if _, ok := FromTrans(ctx); ok {
		return fn(ctx)
	}

	// 创建新的事务并执行
	return a.DB.Transaction(func(db *gorm.DB) error {
		return fn(NewTrans(ctx, db))
	})
}

// GetDB 获取数据库连接实例，用于具体的数据库操作（dal）
func GetDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	db := defDB
	// 如果上下文中存在事务,使用事务的数据库连接（Transaction 的匿名函数参数的 db 参数）
	if tdb, ok := FromTrans(ctx); ok {
		db = tdb
	}
	// 如果需要行锁,添加FOR UPDATE子句
	if FromRowLock(ctx) {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"}) // 排他锁（写锁）
	}
	return db.WithContext(ctx)
}
