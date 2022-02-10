package redis

import (
	"context"

	"github.com/afocus/trace"
	"github.com/go-redis/redis/v8"
)

type Hook struct{}

func (h Hook) BeforeProcess(c context.Context, cmd redis.Cmder) (context.Context, error) {
	e := trace.Start(c, "redis "+cmd.Name(), trace.Attribute("db.system", "redis"))
	return e.WithContext(e.Context()), nil
}

func (h Hook) AfterProcess(c context.Context, cmd redis.Cmder) error {
	e := trace.FromContext(c)
	if e != nil {
		e.End()
	}
	return nil
}

func (h Hook) BeforeProcessPipeline(c context.Context, cmds []redis.Cmder) (context.Context, error) {
	e := trace.Start(c, "redis pipeline", trace.Attribute("db.system", "redis"))
	return e.WithContext(e.Context()), nil

}

func (h Hook) AfterProcessPipeline(c context.Context, cmds []redis.Cmder) error {
	e := trace.FromContext(c)
	if e != nil {
		e.End()
	}
	return nil
}
