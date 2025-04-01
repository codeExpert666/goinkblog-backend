package logging

import (
	"context"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Logger LoggerConfig
}

type LoggerConfig struct {
	Debug      bool   `json:"debug"`
	Level      string `json:"level"` // debug/info/warn/error/dpanic/panic/fatal
	CallerSkip int    `json:"caller_skip"`
	File       struct {
		Enable     bool   `json:"enable"`
		Path       string `json:"path"`
		MaxSize    int    `json:"max_size"`
		MaxBackups int    `json:"max_backups"`
	} `json:"file"`
	Hooks []*HookConfig `json:"hooks"`
}

type HookConfig struct {
	Enable    bool              `json:"enable"`
	Level     string            `json:"level"`
	MaxBuffer int               `json:"max_buffer"`
	MaxThread int               `json:"max_thread"`
	Options   map[string]string `json:"options"` // 数据库配置
	Extra     map[string]string `json:"extra"`   // 日志录入数据库额外添加的信息
}

type HookHandlerFunc func(ctx context.Context, hookCfg *HookConfig) (*Hook, error)

func InitWithConfig(ctx context.Context, cfg *LoggerConfig, hookHandle ...HookHandlerFunc) (func(), error) {
	var zconfig zap.Config
	if cfg.Debug {
		cfg.Level = "debug"
		zconfig = zap.NewDevelopmentConfig()
	} else {
		zconfig = zap.NewProductionConfig()
	}

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	zconfig.Level.SetLevel(level)

	var (
		logger   *zap.Logger
		cleanFns []func()
	)

	if cfg.File.Enable {
		filename := cfg.File.Path
		_ = os.MkdirAll(filepath.Dir(filename), 0777)
		fileWriter := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    cfg.File.MaxSize,
			MaxBackups: cfg.File.MaxBackups,
			Compress:   false,
			LocalTime:  true,
		}

		cleanFns = append(cleanFns, func() {
			_ = fileWriter.Close()
		})

		zc := zapcore.NewCore(
			zapcore.NewJSONEncoder(zconfig.EncoderConfig),
			zapcore.AddSync(fileWriter),
			zconfig.Level,
		)
		logger = zap.New(zc)
	} else {
		ilogger, err := zconfig.Build()
		if err != nil {
			return nil, err
		}
		logger = ilogger
	}

	skip := cfg.CallerSkip
	if skip <= 0 {
		skip = 2
	}

	logger = logger.WithOptions(
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
		zap.AddCallerSkip(skip),
	)

	for _, h := range cfg.Hooks {
		if !h.Enable || len(hookHandle) == 0 {
			continue
		}

		writer, err := hookHandle[0](ctx, h)
		if err != nil {
			return nil, err
		} else if writer == nil {
			continue
		}

		cleanFns = append(cleanFns, func() {
			writer.Flush()
		})

		hookLevel := zap.NewAtomicLevel()
		if level, err := zapcore.ParseLevel(h.Level); err == nil {
			hookLevel.SetLevel(level)
		} else {
			hookLevel.SetLevel(zap.InfoLevel)
		}

		hookEncoder := zap.NewProductionEncoderConfig()
		hookEncoder.EncodeTime = zapcore.EpochMillisTimeEncoder
		hookEncoder.EncodeDuration = zapcore.MillisDurationEncoder
		hookCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(hookEncoder),
			zapcore.AddSync(writer),
			hookLevel,
		)

		logger = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, hookCore)
		}))
	}

	zap.ReplaceGlobals(logger)
	return func() {
		for _, fn := range cleanFns {
			fn()
		}
	}, nil
}
