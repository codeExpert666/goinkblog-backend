package config

import (
	"fmt"

	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
)

// Config 配置参数
type Config struct {
	General    General              `json:"general"`
	Logger     logging.LoggerConfig `json:"logger"`
	Storage    Storage              `json:"storage"`
	Middleware Middleware           `json:"middleware"`
	Util       Util                 `json:"util"`
	Dictionary Dictionary           `json:"dictionary"`
	AI         AIConfig             `json:"ai"`
}

type General struct {
	AppName            string `default:"goinkblog" json:"app_name"`
	Version            string `default:"v1.0.0" json:"version"`
	Debug              bool   `json:"debug"`
	WorkDir            string // 从命令参数获取
	DisableSwagger     bool   `json:"disable_swagger"`
	DisablePrintConfig bool   `json:"disable_print_config"`
	PprofAddr          string `json:"pprof_addr"`

	HTTP struct {
		Addr            string `default:":8080" json:"addr"`
		ShutdownTimeout int    `default:"30" json:"shutdown_timeout"`
		MaxContentLen   int64  `default:"67108864" json:"max_content_length"`
		ReadTimeout     int    `default:"60" json:"read_timeout"`
		WriteTimeout    int    `default:"60" json:"write_timeout"`
		IdleTimeout     int    `default:"10" json:"idle_timeout"`
		CertFile        string `json:"cert_file"`
		KeyFile         string `json:"key_file"`
	} `json:"http"`

	Admin struct {
		ID       uint   `default:"1" json:"id"`
		Username string `default:"XinnZ" json:"username"`
		Email    string `default:"zhouxin23333@gmail.com" json:"email"`
		Password string `json:"password"`
		Avatar   string `default:"/pic/avatars/admin.jpg" json:"avatar"`
		Bio      string `default:"I'm the administrator of GoInkBlog." json:"bio"`
		Role     string `default:"admin" json:"role"`
	} `json:"admin"`
}

type Storage struct {
	DB struct {
		Debug        bool   `json:"debug"`
		AutoMigrate  bool   `json:"auto_migrate"`
		DSN          string `default:"root:123456@tcp(127.0.0.1:3306)/goinkblog?charset=utf8mb4&parseTime=True&loc=Local" json:"dsn"`
		MaxIdleConns int    `default:"50" json:"max_idle_conns"`
		MaxOpenConns int    `default:"100" json:"max_open_conns"`
		MaxLifetime  int    `default:"86400" json:"max_life_time"`
		MaxIdleTime  int    `default:"50" json:"max_idle_time"`
		TablePrefix  string `json:"table_prefix"`
		PrepareStmt  bool   `json:"prepare_stmt"`
	} `json:"db"`

	Cache struct {
		Delimiter string `default:":" json:"delimiter"`
		Redis     struct {
			Addr     string `default:"127.0.0.1:6379" json:"addr"`
			DB       int    `json:"db"`
			Username string `default:"root" json:"username"`
			Password string `default:"123456" json:"password"`
		} `json:"redis"`
	} `json:"cache"`
}

type Util struct {
	Captcha struct {
		Length int `default:"4" json:"length"`   // 验证码长度
		Width  int `default:"400" json:"width"`  // 验证码图片宽度
		Height int `default:"160" json:"height"` // 验证码图片高度
		Redis  struct {
			Addr      string `json:"addr"`                          // Redis服务器地址
			Username  string `json:"username"`                      // Redis用户名
			Password  string `json:"password"`                      // Redis密码
			DB        int    `json:"db"`                            // Redis数据库索引
			KeyPrefix string `default:"captcha:" json:"key_prefix"` // 验证码键前缀
		} `json:"redis"`
	} `json:"captcha"`
}

type AIConfig struct {
	Provider    string  `default:"local" json:"provider"` // openai, local
	APIKey      string  `json:"api_key"`
	Endpoint    string  `default:"http://localhost:11434/api/generate" json:"endpoint"`
	Model       string  `default:"gemma3:12b" json:"model"`
	Temperature float64 `default:"0.7" json:"temperature"`
}

type Dictionary struct {
	UserCacheExp int `default:"4" json:"user_cache_exp"` // 用户缓存过期时间（小时）
}

func (c *Config) IsDebug() bool {
	return c.General.Debug
}

func (c *Config) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		panic("Failed to marshal config: " + err.Error())
	}
	return string(b)
}

func (c *Config) Print() {
	if c.General.DisablePrintConfig {
		return
	}
	fmt.Println("// ----------------------- Load configurations start ------------------------")
	fmt.Println(c.String())
	fmt.Println("// ----------------------- Load configurations end --------------------------")
}

func (c *Config) FormatTableName(name string) string {
	return c.Storage.DB.TablePrefix + name
}

// PreLoad Redis配置自动复用
func (c *Config) PreLoad() {
	addr := c.Storage.Cache.Redis.Addr
	db := c.Storage.Cache.Redis.DB
	username := c.Storage.Cache.Redis.Username
	password := c.Storage.Cache.Redis.Password
	// Redis配置复制到验证码服务
	c.Util.Captcha.Redis.Addr = addr
	c.Util.Captcha.Redis.DB = db
	c.Util.Captcha.Redis.Username = username
	c.Util.Captcha.Redis.Password = password
	// Redis配置复制到限流器
	c.Middleware.RateLimiter.Redis.Addr = addr
	c.Middleware.RateLimiter.Redis.DB = db
	c.Middleware.RateLimiter.Redis.Username = username
	c.Middleware.RateLimiter.Redis.Password = password
	// Redis配置复制到认证服务
	c.Middleware.Auth.Store.Redis.Addr = addr
	c.Middleware.Auth.Store.Redis.DB = db
	c.Middleware.Auth.Store.Redis.Username = username
	c.Middleware.Auth.Store.Redis.Password = password
}
