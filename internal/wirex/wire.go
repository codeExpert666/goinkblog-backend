//go:build wireinject
// +build wireinject

package wirex

// The build tag makes sure the stub is not built in the final build.

import (
	"context"

	"github.com/google/wire"

	"github.com/codeExpert666/goinkblog-backend/internal/mods"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// BuildInjector 构建注入器
func BuildInjector(ctx context.Context) (*Injector, func(), error) {
	wire.Build(
		InitCacher,
		InitDB,
		InitAuth,
		wire.NewSet(wire.Struct(new(util.Trans), "*")),
		wire.NewSet(wire.Struct(new(Injector), "*")),
		mods.Set,
	)
	return new(Injector), nil, nil
}
