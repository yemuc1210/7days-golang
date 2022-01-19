package gee

import (
	"fmt"
	"log"
	"net/http"
)

// HandlerFunc defines the request handler used by gee
// 提供给用户使用 定义路由映射得处理方法
type HandlerFunc func(http.ResponseWriter, *http.Request)

// Engine implement the interface of ServeHTTP
type Engine struct {
	// url : 处理器函数的映射
	router map[string]HandlerFunc
}

// New is the constructor of gee.Engine
func New() *Engine {
	return &Engine{router: make(map[string]HandlerFunc)}
}

// 私有函数：由Get调用
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	key := method + "-" + pattern // 示例：GET-/hello
	log.Printf("Route %4s - %s", method, pattern)
	// 建立映射关系
	engine.router[key] = handler
}

// GET defines the method to add GET request
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST defines the method to add POST request
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

// Run defines the method to start a http server
func (engine *Engine) Run(addr string) (err error) {
	// 对http.ListenAndServe的封装
	// func ListenAndServe(addr string, handler Handler) error
	// handler 是一个处理器示例 里面包含处理逻辑
	return http.ListenAndServe(addr, engine)
}

// 实现Handler 接口   解析请求得路径 查找路由映射表并执行对应方法
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.Method + "-" + req.URL.Path
	if handler, ok := engine.router[key]; ok {
		handler(w, req)
	} else {
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}
