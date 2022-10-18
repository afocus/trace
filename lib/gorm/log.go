package gorm

import (
	"context"
	"time"

	"github.com/afocus/trace"

	"gorm.io/gorm/logger"
)

// Logger 替换gorm的日志
// 主要提供关联trace的日志功能
// param slowsqlshow 显示慢sql语句的最小执行时间
func Logger(slowsqlshow time.Duration) *zerologgerDriver {
	return &zerologgerDriver{
		slowDura: slowsqlshow,
	}
}

// // 实现 gorm.io/gorm/logger 接口
type zerologgerDriver struct {
	slowDura time.Duration
}

// // Trace 主要输出log日志的方法
func (c *zerologgerDriver) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

}

func (c *zerologgerDriver) LogMode(logger.LogLevel) logger.Interface { return c }

func (c *zerologgerDriver) Info(ctx context.Context, s string, v ...interface{}) {
	if e := trace.FromContext(ctx); e != nil {
		e.Log().Info().Msgf(s, v...)
	}
}

func (c *zerologgerDriver) Warn(ctx context.Context, s string, v ...interface{}) {
	if e := trace.FromContext(ctx); e != nil {
		e.Log().Warn().Msgf(s, v...)
	}
}

func (c *zerologgerDriver) Error(ctx context.Context, s string, v ...interface{}) {
	if e := trace.FromContext(ctx); e != nil {
		e.Log().Error().Msgf(s, v...)
	}
}
