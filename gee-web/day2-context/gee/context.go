package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// 起别名 gee.H
type H map[string]interface{}

type Context struct {
	// origin objects  net/http包里的
	Writer http.ResponseWriter
	Req    *http.Request
	// request info   需要关注的请求信息
	Path   string
	Method string
	// response info  需要关注得响应得信息
	StatusCode int
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
	}
}

// 提供给用户调用的函数
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

//
func (c *Context) Query(key string) string {
	// func (u *URL) Query() Values
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	//func (ResponseWriter) Header() Header
	// func (h Header) Set(key string, value string)
	c.Writer.Header().Set(key, value) // 链式调用
}

// 快速构建String类型的响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}
