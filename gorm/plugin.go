package gorm

import (
	"fmt"
	"tempotest/traceing"

	"gorm.io/gorm"
)

type plugin struct{}

func Plugin() *plugin {
	return &plugin{}
}

func (g *plugin) Name() string {
	return "gorm-otel-traceing"
}

func (g *plugin) Initialize(db *gorm.DB) error {

	createCallback := db.Callback().Create()
	createCallback.Before("gorm:create").Register("traceing:before_create", g.before("create"))
	createCallback.After("gorm:create").Register("traceing:after_create", g.after)

	queryCallback := db.Callback().Query()
	queryCallback.Before("gorm:query").Register("traceing:before_query", g.before("query"))
	queryCallback.After("gorm:query").Register("traceing:after_query", g.after)

	deleteCallback := db.Callback().Delete()
	deleteCallback.Before("gorm:delete").Register("traceing:before_delete", g.before("delete"))
	deleteCallback.After("gorm:delete").Register("traceing:after_delete", g.after)

	updateCallback := db.Callback().Update()
	updateCallback.Before("gorm:update").Register("traceing:before_update", g.before("update"))
	updateCallback.After("gorm:update").Register("traceing:after_update", g.after)

	rowCallback := db.Callback().Row()
	rowCallback.Before("gorm:row").Register("traceing:before_row", g.before("row"))
	rowCallback.After("gorm:row").Register("traceing:after_row", g.after)

	rawCallback := db.Callback().Raw()
	rawCallback.Before("gorm:raw").Register("traceing:before_raw", g.before("raw"))
	rawCallback.After("gorm:raw").Register("traceing:after_raw", g.after)

	return nil

}

func (g *plugin) before(method string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Statement.Context == nil {
			return
		}

		e := traceing.Start(
			db.Statement.Context,
			fmt.Sprintf("database %s", method),
			traceing.Attribute("db.sql", db.Statement.SQL.String()),
			traceing.Attribute("db.system", db.Dialector.Name()),
		)

		db.Statement.Context = e.WithContext(e.Context())
	}
}
func (g *plugin) after(db *gorm.DB) {
	if db.Statement.Context == nil {
		return
	}
	if e := traceing.FromContext(db.Statement.Context); e != nil {
		if db.Error != nil {
			e.EndError(db.Error)
		} else {
			e.EndOK()
		}
	}
}
