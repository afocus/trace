package redis

import (
	"context"
	"tempotest/traceing"

	"github.com/go-redis/redis/v8"
)

type Hook struct{}

func (h Hook) BeforeProcess(c context.Context, cmd redis.Cmder) (context.Context, error) {
	e := traceing.Start(c, "redis "+cmd.Name(), traceing.Attribute("db.system", "redis"))
	return e.WithContext(e.Context()), nil
}

func (h Hook) AfterProcess(c context.Context, cmd redis.Cmder) error {
	e := traceing.FromContext(c)
	if e != nil {
		e.End()
	}
	return nil
}

func (h Hook) BeforeProcessPipeline(c context.Context, cmds []redis.Cmder) (context.Context, error) {
	e := traceing.Start(c, "redis pipeline", traceing.Attribute("db.system", "redis"))
	return e.WithContext(e.Context()), nil

}

func (h Hook) AfterProcessPipeline(c context.Context, cmds []redis.Cmder) error {
	e := traceing.FromContext(c)
	if e != nil {
		e.End()
	}
	return nil
}
