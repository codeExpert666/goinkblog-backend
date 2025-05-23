package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/creasty/defaults"
)

var (
	once sync.Once
	C    = new(Config)
)

func MustLoad(dir string, names ...string) {
	once.Do(func() {
		if err := Load(dir, names...); err != nil {
			panic(err)
		}
	})
}

// Load 从指定目录加载配置文件并解析到结构体中
func Load(dir string, names ...string) error {
	if err := defaults.Set(C); err != nil {
		return err
	}

	// parseFile 解析单个配置文件
	parseFile := func(name string) error {
		buf, err := os.ReadFile(name)
		if err != nil {
			return errors.Wrapf(err, "failed to read config file %s", name)
		}
		err = json.Unmarshal(buf, C)
		return errors.Wrapf(err, "failed to unmarshal config %s", name)
	}

	for _, name := range names {
		fullname := filepath.Join(dir, name)
		info, err := os.Stat(fullname)
		if err != nil {
			return errors.Wrapf(err, "failed to get config file %s", name)
		}

		if info.IsDir() {
			err := filepath.WalkDir(fullname, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				} else if d.IsDir() { // 进入目录
					return nil
				}
				return parseFile(path)
			})
			if err != nil {
				return errors.Wrapf(err, "failed to walk config dir %s", name)
			}
			continue
		}
		if err := parseFile(fullname); err != nil {
			return err
		}
	}

	return nil
}
