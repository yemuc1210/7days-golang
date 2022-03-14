package main

// $ curl http://localhost:9999/
// URL.Path = "/"
// $ curl http://localhost:9999/hello
// Header["Accept"] = ["*/*"]
// Header["User-Agent"] = ["curl/7.54.0"]
// curl http://localhost:9999/world
// 404 NOT FOUND: /world

import (
	"fmt"
	"log"
	"net/http"
)

// Engine is the uni handler for all requests
//Engin结构体实现ServeHTTP方法 满足了Handler接口
type Engine struct{}

// HTTP请求会交给该实例处理
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/":
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	case "/hello":
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	default:
		fmt.Fprintf(w, "404 NOT FOUND: %s\n", req.URL)
	}
}

func main() {
	engine := new(Engine)
	// 所有的HTTP请求会转向自定义的处理逻辑了！
	log.Fatal(http.ListenAndServe(":9999", engine))
	/**
	在实现Engine之前，我们调用 http.HandleFunc 实现了路由和Handler的映射，
	也就是只能针对具体的路由写处理逻辑。比如/hello。
	但是在实现Engine之后，我们拦截了所有的HTTP请求，拥有了统一的控制入口。
	在这里我们可以自由定义路由映射的规则，也可以统一添加一些处理逻辑，例如日志、异常处理等。
	*/
}
