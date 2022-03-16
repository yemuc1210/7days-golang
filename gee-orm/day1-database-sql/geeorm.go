package geeorm

import (
	"database/sql"

	"geeorm/log"
	"geeorm/session"
)

// Session 负责与数据库的交互
// Engine 负责交互前的准备工作：连接、测试数据库，关闭连接等
// Engine 负责用户交互
// Engine is the main struct of geeorm, manages all db sessions and transactions.
type Engine struct {
	db *sql.DB
}

// NewEngine create a instance of Engine
// connect database and ping it to test whether it's alive
func NewEngine(driver, source string) (e *Engine, err error) {
	// 连接数据库
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// Send a ping to make sure the database connection is alive.
	// 测试数据库连接是否alive
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	// 返回实例
	e = &Engine{db: db}
	log.Info("Connect database success")
	return
}

// Close database connection
func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("Failed to close database")
	}
	log.Info("Close database success")
}

// NewSession creates a new session for next operations
// 创建session，处理数据库的交互
func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db)
}
