package config

const (
	// CacheNSForUser 用户相关的缓存命名空间
	CacheNSForUser = "user"

	// CacheNSForRole 角色相关的缓存命名空间
	CacheNSForRole = "role"

	// CacheNSForAI AI 模块相关的缓存命名空间
	CacheNSForAI = "ai"
)

const (
	// CacheKeyForSyncToModelSelector 模型选择器同步标记的缓存键
	CacheKeyForSyncToModelSelector = "sync:model_selector"

	// CacheKeyForSyncToCasbin Casbin同步标记的缓存键
	CacheKeyForSyncToCasbin = "sync:casbin"
)

const (
	// SupportedImageFormats 系统支持的图片格式
	SupportedImageFormats = ".jpg, .jpeg, .png, .bmp, .webp"
)
