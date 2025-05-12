package config

type Middleware struct {
	CORS struct {
		Enable                 bool     `json:"enable"`
		AllowAllOrigins        bool     `json:"allow_all_origins"`
		AllowOrigins           []string `json:"allow_origins"`
		AllowMethods           []string `json:"allow_methods"`
		AllowHeaders           []string `json:"allow_headers"`
		AllowCredentials       bool     `json:"allow_credentials"`
		ExposeHeaders          []string `json:"expose_headers"`
		MaxAge                 int      `json:"max_age"`
		AllowWildcard          bool     `json:"allow_wildcard"`
		AllowBrowserExtensions bool     `json:"allow_browser_extensions"`
		AllowWebSockets        bool     `json:"allow_web_sockets"`
		AllowFiles             bool     `json:"allow_files"`
	} `json:"cors"`

	Auth struct {
		Disable             bool     `json:"disable"`
		SkippedPathPrefixes []string `json:"skipped_path_prefixes"`
		SigningMethod       string   `default:"HS512" json:"signing_method"`
		SigningKey          string   `default:"fqe@#$%^fa^&&*()_+" json:"signing_key"`
		OldSigningKey       string   `json:"old_signing_key"`
		Expired             int      `default:"86400" json:"expired"`

		Store struct {
			Delimiter string `default:":" json:"delimiter"`
			Redis     struct {
				Addr     string `json:"addr"`
				DB       int    `json:"db"`
				Username string `json:"username"`
				Password string `json:"password"`
			}
		} `json:"store"`
	} `json:"auth"`

	Trace struct {
		SkippedPathPrefixes []string `json:"skipped_path_prefixes"`
		RequestHeaderKey    string   `default:"X-Request-Id" json:"request_header_key"`
		ResponseTraceKey    string   `default:"X-Trace-Id" json:"response_trace_key"`
	} `json:"trace"`

	Logger struct {
		SkippedPathPrefixes      []string `json:"skipped_path_prefixes"`
		MaxOutputRequestBodyLen  int      `default:"4096" json:"max_output_request_body_len"`
		MaxOutputResponseBodyLen int      `default:"1024" json:"max_output_response_body_len"`
	} `json:"logger"`

	CopyBody struct {
		SkippedPathPrefixes []string `json:"skipped_path_prefixes"`
		MaxContentLen       int64    `default:"67108864" json:"max_content_len"`
	} `json:"copy_body"`

	Recovery struct {
		Skip int `default:"2" json:"skip"`
	} `json:"recovery"`

	RateLimiter struct {
		Enable              bool     `json:"enable"`
		SkippedPathPrefixes []string `json:"skipped_path_prefixes"`
		IPLimit             int      `json:"ip_limit"`
		UserLimit           int      `json:"user_limit"`
		Redis               struct {
			Addr     string `json:"addr"`
			DB       int    `json:"db"`
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"redis"`
	} `json:"rate_limiter"`

	Casbin struct {
		SkippedPathPrefixes []string `json:"skipped_path_prefixes"`
		ModelFile           string   `default:"rbac_model.conf" json:"model_file"`  // RBAC 模型配置文件路径
		PolicyFile          string   `default:"rbac_policy.csv" json:"policy_file"` // RBAC 策略配置文件路径
		AutoLoadInterval    int      `json:"auto_load_interval"`                    // 秒
	} `json:"casbin"`

	Static struct { // 命令行参数
		Dir string `json:"dir"`
	} `json:"static"`
}
