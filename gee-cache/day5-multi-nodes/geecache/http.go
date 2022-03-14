package geecache

import (
	"fmt"
	"geecache/consistenthash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	// 默认虚拟节点倍数
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
// 实现PeerPicker接口，根据传入的key值，选择相应的节点，然后读取其缓存
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	// 自身
	self     string
	basePath string
	// 互斥锁
	mu sync.Mutex // guards peers and httpGetters
	// 一致性哈希  通过一致性哈希+key选择合适的peers
	peers *consistenthash.Map
	// Http 客户端类 需要实现PeerGetter接口
	// 每一个远程节点对应一个httpGetter，而httpGetter与远程节点的地址baseURL有关
	httpGetters map[string]*httpGetter // key-ed by e.g. "http://10.0.0.2:8008"
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	// 但是mu peers  httpGetters怎么没有初始化
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
// 实现handler接口，从而保证HTTPPool可以处理http请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	// 根据group名 获取group实例
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	// 从 group里面根据key 获取缓存
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建消息
	// 首先构建消息头
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		return
	}
}

// Set updates the pool's list of peers.
// HTTPPool 的set操作，更新peers列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 一致性哈希的New操作：常见一个Map实例
	// fn  支持自定义的哈希函数
	p.peers = consistenthash.New(defaultReplicas, nil)
	// 将peers传入一致性哈希，每个节点需要构建虚拟节点，并通过字典将虚拟节点和真实节点绑定
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	// 每个peer节点对应一个httpGetter
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
// 实现 PickPeer 接口，根据key值选择一个peer
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 将key传入一致性哈希环，会返回一个合适的对等节点
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		// 然后查找对等节点的PeerGetter，从而获得具体的数据
		return p.httpGetters[peer], true
	}
	return nil, false
}

// 通过强制类型转换，由编译器检查 HTTPool是否实现了PeerPicker接口
var _ PeerPicker = (*HTTPPool)(nil)

// httpGetter  客户端类  需要实现PeerGetter接口
type httpGetter struct {
	baseURL string
}

// Get PeerGetter 接口的Get函数
// 从某个group里面查找key相关的缓存值
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	// /_geecachegroup/key
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	// 发起一个http Get 请求，由相应的http handler处理
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("httpGetter Get Body.Close() Error")
		}
	}(res.Body)

	// 先查看头部状态码
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	// ok  读取消息体
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

// 强制类型转换
// 由编译器判断 httpGetter 是否实现了PeerGetter接口
var _ PeerGetter = (*httpGetter)(nil)
