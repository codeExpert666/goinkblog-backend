package schema

// SystemInfo 系统信息
type SystemInfo struct {
	// 系统运行时间（秒）
	Uptime int64 `json:"uptime"`
	// 主机名
	Hostname string `json:"hostname"`
	// 平台信息
	Platform string `json:"platform"`
	// 平台版本
	PlatformVersion string `json:"platform_version"`
	// 内核版本
	KernelVersion string `json:"kernel_version"`
	// 系统架构
	Architecture string `json:"architecture"`
	// 进程数量
	ProcessCount int `json:"process_count"`
}

// CPUInfo CPU使用情况
type CPUInfo struct {
	// CPU核心数
	Cores int `json:"cores"`
	// CPU使用率（百分比）
	UsagePercent float64 `json:"usage_percent"`
	// CPU型号信息
	ModelName string `json:"model_name"`
	// 每个CPU核心的使用率
	PerCorePercent []float64 `json:"per_core_percent"`
	// CPU频率（MHz）
	Frequency float64 `json:"frequency"`
	// CPU缓存大小（字节）
	CacheSize int32 `json:"cache_size"`
}

// MemoryInfo 内存使用情况
type MemoryInfo struct {
	// 总内存（字节）
	Total uint64 `json:"total"`
	// 已使用内存（字节）
	Used uint64 `json:"used"`
	// 可用内存（字节）
	Available uint64 `json:"available"`
	// 内存使用率（百分比）
	UsagePercent float64 `json:"usage_percent"`
	// 交换分区总大小（字节）
	SwapTotal uint64 `json:"swap_total"`
	// 交换分区已用（字节）
	SwapUsed uint64 `json:"swap_used"`
	// 可用交换分区（字节）
	SwapFree uint64 `json:"swap_free"`
	// 交换分区使用率（百分比）
	SwapUsagePercent float64 `json:"swap_usage_percent"`
}

// DiskInfo 磁盘使用情况
type DiskInfo struct {
	// 磁盘分区信息
	Partitions []PartitionInfo `json:"partitions"`
	// 自系统启动以来的读取操作总次数
	IoCountersReadCount uint64 `json:"io_counters_read_count"`
	// 自系统启动以来的写入操作总次数
	IoCountersWriteCount uint64 `json:"io_counters_write_count"`
	// 自系统启动以来读取的总字节数
	IoCountersReadBytes uint64 `json:"io_counters_read_bytes"`
	// 自系统启动以来写入的总字节数
	IoCountersWriteBytes uint64 `json:"io_counters_write_bytes"`
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	// 设备名称
	Device string `json:"device"`
	// 挂载点
	Mountpoint string `json:"mountpoint"`
	// 文件系统类型
	Fstype string `json:"fstype"`
	// 总空间（字节）
	Total uint64 `json:"total"`
	// 已使用空间（字节）
	Used uint64 `json:"used"`
	// 可用空间（字节）
	Free uint64 `json:"free"`
	// 使用率（百分比）
	UsagePercent float64 `json:"usage_percent"`
}

// GoRuntimeInfo Go运行时信息
type GoRuntimeInfo struct {
	// Go版本
	Version string `json:"version"`
	// 当前goroutine数量
	Goroutines int `json:"goroutines"`
	// 自程序启动以来GC次数
	GCCount uint32 `json:"gc_count"`
	// 最近一次GC暂停时间（纳秒）
	GCPauseNs uint64 `json:"gc_pause_ns"`
	// 堆对象数量
	HeapObjects uint64 `json:"heap_objects"`
	// 堆分配字节数
	HeapAlloc uint64 `json:"heap_alloc"`
}

// DatabaseInfo 数据库状态信息
type DatabaseInfo struct {
	// 数据库类型
	DBType string `json:"db_type"`
	// 数据库版本
	Version string `json:"version"`
	// 连接池状态
	Pool PoolStatus `json:"pool"`
	// 慢查询阈值（秒）
	SlowQueryThreshold float64 `json:"slow_query_threshold"`
	// 慢查询数量（最近1小时）
	SlowQueryCount int `json:"slow_query_count"`
	// 活跃事务数量
	ActiveTransactions int `json:"active_transactions"`
	// 数据库运行时间（秒）
	Uptime int64 `json:"uptime"`
	// 每秒查询数
	QPS float64 `json:"qps"`
	// 缓存命中率（百分比）
	CacheHitRate float64 `json:"cache_hit_rate"`
	// 数据库大小（字节）
	DBSize int64 `json:"db_size"`
	// 表数量
	TableCount int `json:"table_count"`
	// 索引数量
	IndexCount int `json:"index_count"`
}

// PoolStatus 连接池状态
type PoolStatus struct {
	// 最大连接数
	MaxOpen int `json:"max_open"`
	// 最大空闲连接数
	MaxIdle int `json:"max_idle"`
	// 活跃连接数
	Open int `json:"open"`
	// 空闲连接数
	Idle int `json:"idle"`
	// 使用中的连接数
	InUse int `json:"in_use"`
	// 等待连接数
	WaitCount int64 `json:"wait_count"`
	// 连接生命周期（秒）
	MaxLifetime int `json:"max_lifetime"`
	// 最大空闲时间（秒）
	MaxIdleTime int `json:"max_idle_time"`
}

// CacheInfo Redis缓存状态信息
type CacheInfo struct {
	// 缓存类型
	CacheType string `json:"cache_type"`
	// 缓存版本
	Version string `json:"version"`
	// 已连接客户端数量
	ConnectedClients int64 `json:"connected_clients"`
	// 已使用内存（字节）
	UsedMemory int64 `json:"used_memory"`
	// 已使用内存峰值（字节）
	UsedMemoryPeak int64 `json:"used_memory_peak"`
	// 命中率
	HitRate float64 `json:"hit_rate"`
	// 键数量
	KeyCount int64 `json:"key_count"`
}
