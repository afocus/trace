package gorm

import (
	"fmt"

	"github.com/afocus/trace"

	"gorm.io/gorm"
)

type plugin struct{}

func Plugin() *plugin {
	return &plugin{}
}

func (g *plugin) Name() string {
	return "gorm-otel-trace"
}

func (g *plugin) Initialize(db *gorm.DB) error {

	createCallback := db.Callback().Create()
	createCallback.Before("gorm:create").Register("trace:before_create", g.before("create"))
	createCallback.After("gorm:create").Register("trace:after_create", g.after)

	queryCallback := db.Callback().Query()
	queryCallback.Before("gorm:query").Register("trace:before_query", g.before("query"))
	queryCallback.After("gorm:query").Register("trace:after_query", g.after)

	deleteCallback := db.Callback().Delete()
	deleteCallback.Before("gorm:delete").Register("trace:before_delete", g.before("delete"))
	deleteCallback.After("gorm:delete").Register("trace:after_delete", g.after)

	updateCallback := db.Callback().Update()
	updateCallback.Before("gorm:update").Register("trace:before_update", g.before("update"))
	updateCallback.After("gorm:update").Register("trace:after_update", g.after)

	rowCallback := db.Callback().Row()
	rowCallback.Before("gorm:row").Register("trace:before_row", g.before("row"))
	rowCallback.After("gorm:row").Register("trace:after_row", g.after)

	rawCallback := db.Callback().Raw()
	rawCallback.Before("gorm:raw").Register("trace:before_raw", g.before("raw"))
	rawCallback.After("gorm:raw").Register("trace:after_raw", g.after)

	return nil

}

func (g *plugin) before(method string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement.Context == nil {
			return
		}
		name := fmt.Sprintf("database %s", method)
		_, ctx := trace.Start(db.Statement.Context, name, trace.Attribute("db.system", db.Dialector.Name()))
		db.Statement.Context = ctx
	}
}
func (g *plugin) after(db *gorm.DB) {
	if db.Statement.Context == nil {
		return
	}
	if e := trace.FromContext(db.Statement.Context); e != nil {
		e.End(db.Error)
	}
}
