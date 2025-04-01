package wirex

import (
	"context"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
	"github.com/codeExpert666/goinkblog-backend/pkg/gormx"
	"github.com/codeExpert666/goinkblog-backend/pkg/jwtx"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// Injector 注入器
type Injector struct {
	DB    *gorm.DB
	Cache cachex.Cacher
	Auth  jwtx.Auther
	M     *mods.Mods
}

// InitDB 初始化数据库
func InitDB(ctx context.Context) (*gorm.DB, func(), error) {
	cfg := config.C.Storage.DB

	db, err := gormx.New(gormx.Config{
		Debug:        cfg.Debug,
		PrepareStmt:  cfg.PrepareStmt,
		DSN:          cfg.DSN,
		MaxLifetime:  cfg.MaxLifetime,
		MaxIdleTime:  cfg.MaxIdleTime,
		MaxOpenConns: cfg.MaxOpenConns,
		MaxIdleConns: cfg.MaxIdleConns,
		TablePrefix:  cfg.TablePrefix,
	})
	if err != nil {
		return nil, nil, err
	}

	return db, func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}, nil
}

// InitCacher 初始化缓存
func InitCacher(ctx context.Context) (cachex.Cacher, func(), error) {
	cfg := config.C.Storage.Cache

	cache := cachex.NewRedisCache(cachex.RedisConfig{
		Addr:     cfg.Redis.Addr,
		DB:       cfg.Redis.DB,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	}, cachex.WithDelimiter(cfg.Delimiter))

	return cache, func() {
		_ = cache.Close(ctx)
	}, nil
}

// InitAuth 初始化认证
func InitAuth(ctx context.Context) (jwtx.Auther, func(), error) {
	cfg := config.C.Middleware.Auth
	var opts []jwtx.Option
	opts = append(opts, jwtx.SetExpired(cfg.Expired))
	opts = append(opts, jwtx.SetSigningKey(cfg.SigningKey, cfg.OldSigningKey))

	var method jwt.SigningMethod
	switch cfg.SigningMethod {
	case "HS256":
		method = jwt.SigningMethodHS256
	case "HS384":
		method = jwt.SigningMethodHS384
	default:
		method = jwt.SigningMethodHS512
	}
	opts = append(opts, jwtx.SetSigningMethod(method))

	cache := cachex.NewRedisCache(cachex.RedisConfig{
		Addr:     cfg.Store.Redis.Addr,
		DB:       cfg.Store.Redis.DB,
		Username: cfg.Store.Redis.Username,
		Password: cfg.Store.Redis.Password,
	}, cachex.WithDelimiter(cfg.Store.Delimiter))

	auth := jwtx.New(jwtx.NewStoreWithCache(cache), opts...)
	return auth, func() {
		_ = auth.Release(ctx)
	}, nil
}
