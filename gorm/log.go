package gorm

import (
	"context"
	"tempotest/traceing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	duration := time.Since(begin)
	if err == nil && duration <= c.slowDura {
		return
	}
	var logevt *zerolog.Event
	if e := traceing.FromContext(ctx); e != nil {
		if err != nil {
			logevt = e.Log().Error().Err(err)
		} else {
			logevt = e.Log().Info()
		}
	} else {
		if err != nil {
			logevt = log.Error().Err(err)
		} else {
			logevt = log.Info()
		}
	}
	sql, rows := fc()
	logevt.Str("sql", sql).Int64("rows", rows).Dur("cost", time.Since(begin)).Send()
}

func (c *zerologgerDriver) LogMode(logger.LogLevel) logger.Interface { return c }

func (c *zerologgerDriver) Info(context.Context, string, ...interface{}) {}

func (c *zerologgerDriver) Warn(context.Context, string, ...interface{}) {}

func (c *zerologgerDriver) Error(context.Context, string, ...interface{}) {}
