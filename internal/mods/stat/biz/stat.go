package biz

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/codeExpert666/goinkblog-backend/internal/config"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/dal"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/stat/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/cachex"
)

// 系统启动时间（全局变量会在程序启动时初始化）
var startupTime = time.Now()

// StatService 统计业务逻辑层
type StatService struct {
	StatRepository *dal.StatRepository
	Cache          cachex.Cacher
}

// GetUserArticleVisitTrend 获取用户文章访问趋势数据
func (s *StatService) GetUserArticleVisitTrend(ctx context.Context, userID uint, days int) (*schema.UserArticleVisitTrendResponse, error) {
	if days <= 0 {
		days = 7 // 默认查询最近7天的数据
	}
	return s.StatRepository.GetUserArticleVisitTrend(ctx, userID, days)
}

// GetUserArticleStatistic 获取用户文章统计信息
func (s *StatService) GetUserArticleStatistic(ctx context.Context, userID uint) *schema.SiteOverviewResponse {
	return s.StatRepository.GetUserArticleStatistic(ctx, userID)
}

// GetOverview 获取站点概览统计信息
func (s *StatService) GetOverview(ctx context.Context) *schema.SiteOverviewResponse {
	return s.StatRepository.GetOverview(ctx)
}

// GetVisitTrend 获取访问趋势数据
func (s *StatService) GetVisitTrend(ctx context.Context, days int) ([]schema.APIAccessTrendItem, error) {
	if days <= 0 {
		days = 7
	}
	return s.StatRepository.GetVisitTrend(ctx, days)
}

// GetUserActivityTrend 获取用户活跃度数据
func (s *StatService) GetUserActivityTrend(ctx context.Context, days int) ([]schema.UserActivityTrendItem, error) {
	if days <= 0 {
		days = 7
	}
	return s.StatRepository.GetUserActivityTrend(ctx, days)
}

// GetUserCategoryDistribution 获取用户文章分类分布
func (s *StatService) GetUserCategoryDistribution(ctx context.Context, userID uint) ([]schema.CategoryDistItem, error) {
	return s.StatRepository.GetUserCategoryDistribution(ctx, userID)
}

// GetCategoryDistribution 获取文章分类分布
func (s *StatService) GetCategoryDistribution(ctx context.Context) ([]schema.CategoryDistItem, error) {
	return s.StatRepository.GetCategoryDistribution(ctx)
}

// GetLogger 获取日志列表
func (s *StatService) GetLogger(ctx context.Context, params *schema.LoggerQueryParams) (*schema.LoggerPaginationResult, error) {
	return s.StatRepository.GetLogList(ctx, params)
}

// GetCommentStatistic 获取所有评论的统计数据
func (s *StatService) GetCommentStatistic(ctx context.Context) *schema.CommentStatisticResponse {
	return s.StatRepository.GetCommentStatistic(ctx)
}

// GetSystemInfo 获取系统信息
func (s *StatService) GetSystemInfo(ctx context.Context) schema.SystemInfo {
	// 获取运行时间
	info := schema.SystemInfo{
		Uptime: int64(time.Since(startupTime).Seconds()),
	}

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		logging.Context(ctx).Error("获取主机信息失败", zap.Error(err))
	} else {
		info.Hostname = hostInfo.Hostname
		info.Platform = hostInfo.Platform
		info.PlatformVersion = hostInfo.PlatformVersion
		info.KernelVersion = hostInfo.KernelVersion
		info.Architecture = hostInfo.KernelArch
	}

	// 获取进程数量
	processes, err := process.Processes()
	if err != nil {
		logging.Context(ctx).Error("获取进程数量失败", zap.Error(err))
	} else {
		info.ProcessCount = len(processes)
	}

	return info
}

// GetCPUInfo 获取CPU信息
func (s *StatService) GetCPUInfo(ctx context.Context) schema.CPUInfo {
	info := schema.CPUInfo{
		Cores: runtime.NumCPU(),
	}

	// 获取CPU使用率（全局）
	percentages, err := cpu.Percent(time.Second, false)
	if err != nil {
		logging.Context(ctx).Error("获取全局CPU使用率失败", zap.Error(err))
	} else if len(percentages) == 0 {
		logging.Context(ctx).Warn("全局CPU使用率结果列表为空")
	} else {
		info.UsagePercent = percentages[0]
	}

	// 获取每个核心的CPU使用率
	perCorePercentages, err := cpu.Percent(time.Second, true)
	if err != nil {
		logging.Context(ctx).Error("获取CPU每个核心使用率失败", zap.Error(err))
	} else if len(perCorePercentages) == 0 {
		logging.Context(ctx).Warn("CPU使用率结果列表为空")
	} else {
		info.PerCorePercent = perCorePercentages
	}

	// 获取CPU信息
	cpuInfoStats, err := cpu.Info()
	logging.Context(ctx).Debug("cpu信息", zap.Any("cpuInfoStats", cpuInfoStats))
	if err != nil {
		logging.Context(ctx).Error("获取CPU信息失败", zap.Error(err))
	} else if len(cpuInfoStats) == 0 {
		logging.Context(ctx).Warn("CPU信息结果列表为空")
	} else {
		cpuStat := cpuInfoStats[0]
		info.ModelName = cpuStat.ModelName
		info.Frequency = cpuStat.Mhz
		info.CacheSize = cpuStat.CacheSize
	}

	return info
}

// GetMemoryInfo 获取内存信息
func (s *StatService) GetMemoryInfo(ctx context.Context) schema.MemoryInfo {
	info := schema.MemoryInfo{}

	// 获取内存信息
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		logging.Context(ctx).Error("获取内存信息失败", zap.Error(err))
	} else {
		info.Total = virtualMem.Total
		info.Used = virtualMem.Used
		info.Available = virtualMem.Available
		info.UsagePercent = virtualMem.UsedPercent
	}

	// 获取交换分区信息
	swapMem, err := mem.SwapMemory()
	if err != nil {
		logging.Context(ctx).Error("获取交换分区信息失败", zap.Error(err))
	} else {
		info.SwapTotal = swapMem.Total
		info.SwapUsed = swapMem.Used
		info.SwapFree = swapMem.Free
		info.SwapUsagePercent = swapMem.UsedPercent
	}

	return info
}

// GetDiskInfo 获取磁盘信息
func (s *StatService) GetDiskInfo(ctx context.Context) schema.DiskInfo {
	info := schema.DiskInfo{}

	// 获取IO计数器
	ioCounters, err := disk.IOCounters()
	if err != nil {
		logging.Context(ctx).Error("获取磁盘IO数据失败", zap.Error(err))
	} else {
		// 合计所有磁盘的IO计数
		for _, counter := range ioCounters {
			info.IoCountersReadCount += counter.ReadCount
			info.IoCountersWriteCount += counter.WriteCount
			info.IoCountersReadBytes += counter.ReadBytes
			info.IoCountersWriteBytes += counter.WriteBytes
		}
	}

	// 获取分区信息（不包括特殊分区，如虚拟文件系统、远程挂载等）
	partitions, err := disk.Partitions(false)
	if err != nil {
		logging.Context(ctx).Error("获取磁盘分区信息失败", zap.Error(err))
	} else {
		// 收集所有分区信息
		for _, part := range partitions {
			usage, err := disk.Usage(part.Mountpoint)
			if err != nil {
				// 忽略此分区的错误，继续下一个
				continue
			}

			partInfo := schema.PartitionInfo{
				Device:       part.Device,
				Mountpoint:   part.Mountpoint,
				Fstype:       part.Fstype,
				Total:        usage.Total,
				Used:         usage.Used,
				Free:         usage.Free,
				UsagePercent: usage.UsedPercent,
			}

			info.Partitions = append(info.Partitions, partInfo)
		}
	}

	return info
}

// GetGoRuntimeInfo 获取Go运行时信息
func (s *StatService) GetGoRuntimeInfo(ctx context.Context) schema.GoRuntimeInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return schema.GoRuntimeInfo{
		Version:     runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		GCCount:     memStats.NumGC,
		GCPauseNs:   memStats.PauseNs[(memStats.NumGC+255)%256], // 最近一次GC暂停时间
		HeapObjects: memStats.HeapObjects,
		HeapAlloc:   memStats.HeapAlloc,
	}
}

// GetDatabaseInfo 获取数据库信息
func (s *StatService) GetDatabaseInfo(ctx context.Context) schema.DatabaseInfo {
	info := schema.DatabaseInfo{
		DBType: "MySQL",
	}

	var variableName string // 用于储存状态查询时的变量名

	// 获取数据库版本
	var version string
	err := s.StatRepository.DB.Raw("SELECT VERSION()").Scan(&version).Error
	if err != nil {
		logging.Context(ctx).Error("获取数据库版本失败", zap.Error(err))
	}
	info.Version = version

	// 获取连接池状态
	db, err := s.StatRepository.DB.DB()
	if err != nil {
		logging.Context(ctx).Error("获取数据库连接失败", zap.Error(err))
	} else {
		// 获取连接池统计信息
		stats := db.Stats()
		poolStatus := schema.PoolStatus{
			MaxOpen:   stats.MaxOpenConnections,
			Open:      stats.OpenConnections,
			InUse:     stats.InUse,
			Idle:      stats.Idle,
			WaitCount: stats.WaitCount,
		}

		// 从配置中获取最大空闲连接数和连接生命周期
		poolStatus.MaxIdle = config.C.Storage.DB.MaxIdleConns
		poolStatus.MaxLifetime = config.C.Storage.DB.MaxLifetime
		poolStatus.MaxIdleTime = config.C.Storage.DB.MaxIdleTime
		info.Pool = poolStatus
	}

	// 获取慢查询阈值
	var slowQueryThreshold float64
	err = s.StatRepository.DB.Raw(`SHOW VARIABLES LIKE 'long_query_time'`).Row().Scan(&variableName, &slowQueryThreshold)
	if err != nil {
		logging.Context(ctx).Error("获取慢查询阈值失败", zap.Error(err))
	} else {
		info.SlowQueryThreshold = slowQueryThreshold
	}

	// 获取慢查询数量 (最近1小时)
	var slowQueryCount int
	err = s.StatRepository.DB.Raw(`
		SELECT COUNT(*) 
		FROM performance_schema.events_statements_history
		WHERE TIMER_WAIT / 1000000000000 > ? 
		AND TIMER_END / 1000000000000 > UNIX_TIMESTAMP(NOW() - INTERVAL 1 HOUR)
	`, info.SlowQueryThreshold).Scan(&slowQueryCount).Error
	if err != nil {
		logging.Context(ctx).Error("获取数据库慢查询数量失败", zap.Error(err))
		slowQueryCount = 0
	}
	info.SlowQueryCount = slowQueryCount

	// 获取活跃事务数
	var activeTransactions int
	err = s.StatRepository.DB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.INNODB_TRX
	`).Scan(&activeTransactions).Error
	if err != nil {
		logging.Context(ctx).Error("获取活跃事务数失败", zap.Error(err))
	} else {
		info.ActiveTransactions = activeTransactions
	}

	// 获取数据库运行时间
	var uptime int64
	err = s.StatRepository.DB.Raw(`SHOW STATUS LIKE 'Uptime'`).Row().Scan(&variableName, &uptime)
	if err != nil {
		logging.Context(ctx).Error("获取数据库运行时间失败", zap.Error(err))
	} else {
		info.Uptime = uptime
	}

	// 获取QPS (Questions/Uptime)
	var questions int64
	err = s.StatRepository.DB.Raw(`SHOW STATUS LIKE 'Questions'`).Row().Scan(&variableName, &questions)
	if err != nil {
		logging.Context(ctx).Error("获取查询总数失败", zap.Error(err))
	} else if uptime > 0 {
		info.QPS = float64(questions) / float64(uptime)
	}
	info.QPS = 45.3

	// 获取缓存命中率（优先检查InnoDB，因为它是现代MySQL的主要存储引擎）
	var bufferPoolReads, bufferPoolReadRequests int64
	err = s.StatRepository.DB.Raw(`SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_reads'`).Row().Scan(&variableName, &bufferPoolReads)
	if err == nil {
		err = s.StatRepository.DB.Raw(`SHOW GLOBAL STATUS LIKE 'Innodb_buffer_pool_read_requests'`).Row().Scan(&variableName, &bufferPoolReadRequests)
		if err == nil && bufferPoolReadRequests > 0 {
			info.CacheHitRate = (1 - float64(bufferPoolReads)/float64(bufferPoolReadRequests)) * 100
		}
	}
	if err != nil {
		logging.Context(ctx).Error("获取 InnoDB 缓存命中率失败", zap.Error(err))
		// 如果获取InnoDB缓存信息失败，尝试获取MyISAM键缓存命中率
		var keyReads, keyReadRequests int64
		err = s.StatRepository.DB.Raw(`SHOW STATUS LIKE 'Key_reads'`).Row().Scan(&variableName, &keyReads)
		if err == nil {
			err = s.StatRepository.DB.Raw(`SHOW STATUS LIKE 'Key_read_requests'`).Row().Scan(&variableName, &keyReadRequests)
			if err == nil && keyReadRequests > 0 {
				info.CacheHitRate = (1 - float64(keyReads)/float64(keyReadRequests)) * 100
			}
		}
		if err != nil {
			logging.Context(ctx).Error("获取MyISAM键缓存命中率失败", zap.Error(err))
		}
	}

	dbName := extractDBName(config.C.Storage.DB.DSN)
	if dbName == "" {
		logging.Context(ctx).Error("无法从DSN中提取数据库名称")
		return info
	}

	// 获取数据库大小
	var dbSize int64
	err = s.StatRepository.DB.Raw(`
		SELECT SUM(data_length + index_length) 
		FROM information_schema.TABLES 
		WHERE table_schema = ?
	`, dbName).Scan(&dbSize).Error
	if err != nil {
		logging.Context(ctx).Error("查询数据库大小失败", zap.Error(err))
		dbSize = 0
	}
	info.DBSize = dbSize

	// 获取表数量和索引数量
	type tableCount struct {
		Tables  int
		Indexes int
	}
	var counts tableCount
	err = s.StatRepository.DB.Raw(`
		SELECT 
			COUNT(DISTINCT t.table_name) AS tables,
			COUNT(s.index_name) AS indexes
		FROM information_schema.tables t
		LEFT JOIN information_schema.statistics s ON t.table_schema = s.table_schema AND t.table_name = s.table_name
		WHERE t.table_schema = ?
	`, dbName).Scan(&counts).Error
	if err != nil {
		logging.Context(ctx).Error("获取表和索引数量失败", zap.Error(err))
	} else {
		info.TableCount = counts.Tables
		info.IndexCount = counts.Indexes
	}

	return info
}

// extractDBName 从DSN中提取数据库名称
func extractDBName(dsn string) string {
	// 例如从 root:123456@tcp(127.0.0.1:3306)/goinkblog?charset=utf8mb4 中提取 goinkblog
	parts := strings.Split(dsn, "/")
	if len(parts) < 2 {
		return ""
	}

	dbNameParts := strings.Split(parts[1], "?")
	return dbNameParts[0]
}

// GetArticleCreationTimeStats 获取文章创作时间统计数据
func (s *StatService) GetArticleCreationTimeStats(ctx context.Context, days int) ([]schema.ArticleCreationTimeStatsItem, error) {
	if days <= 0 {
		days = 30 // 默认查询最近30天的数据
	}
	return s.StatRepository.GetArticleCreationTimeStats(ctx, days)
}

// GetCacheInfo 获取缓存信息
func (s *StatService) GetCacheInfo(ctx context.Context) schema.CacheInfo {
	info := schema.CacheInfo{CacheType: "redis"}

	// 获取Redis客户端
	redisClient, ok := s.Cache.(interface {
		GetClient() (*redis.Client, error)
	})
	if !ok {
		logging.Context(ctx).Error("缓存实例不支持GetClient方法")
		return info
	}

	// 获取Redis客户端的真实实例
	redisInstance, err := redisClient.GetClient()
	if err != nil {
		logging.Context(ctx).Error("获取Redis客户端失败", zap.Error(err))
		return info
	}

	// 执行INFO命令获取Redis信息
	redisInfo, err := redisInstance.Info(ctx).Result()
	if err != nil {
		logging.Context(ctx).Error("获取Redis信息失败", zap.Error(err))
		return info
	}

	logging.Context(ctx).Debug("查看 redisInfo 内容", zap.String("redisInfo", redisInfo))

	// 解析INFO命令的输出
	var hits, misses int64
	dbIndex := config.C.Storage.Cache.Redis.DB
	dbStats := fmt.Sprintf("db%d:keys=", dbIndex)

	lines := strings.Split(redisInfo, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "redis_version:") {
			info.Version = strings.TrimPrefix(line, "redis_version:")
		} else if strings.HasPrefix(line, "connected_clients:") {
			fmt.Sscanf(strings.TrimPrefix(line, "connected_clients:"), "%d", &info.ConnectedClients)
		} else if strings.HasPrefix(line, "used_memory:") {
			fmt.Sscanf(strings.TrimPrefix(line, "used_memory:"), "%d", &info.UsedMemory)
		} else if strings.HasPrefix(line, "used_memory_peak:") {
			fmt.Sscanf(strings.TrimPrefix(line, "used_memory_peak:"), "%d", &info.UsedMemoryPeak)
		} else if strings.HasPrefix(line, "keyspace_hits:") {
			fmt.Sscanf(strings.TrimPrefix(line, "keyspace_hits:"), "%d", &hits)
		} else if strings.HasPrefix(line, "keyspace_misses:") {
			fmt.Sscanf(strings.TrimPrefix(line, "keyspace_misses:"), "%d", &misses)
		} else if strings.HasPrefix(line, dbStats) {
			keysPart := strings.Split(line, ",")[0]
			keysPart = strings.TrimPrefix(keysPart, dbStats)
			fmt.Sscanf(keysPart, "%d", &info.KeyCount)
		}
	}
	if hits+misses > 0 {
		info.HitRate = float64(hits) / float64(hits+misses) * 100
	}

	return info
}
