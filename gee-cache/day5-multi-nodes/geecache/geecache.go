package geecache

import (
	"fmt"
	"log"
	"sync"
)

// A Group is a cache namespace and associated data loaded spread over
// 缓存命名空间，是与用户交互的接口
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	// 新增加peers实例，记录其附近的对等节点
	peers PeerPicker
}

// Getter 加载key对应的数据到缓存中
// 具体如何获得数据，由用户决定
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 为函数类型 GetterFunc
// GetterFunc 使用一个函数实现Getter
// GetterFunc 是接口型函数
// 接口型函数：适用于接口只含有一个方法的情况
type GetterFunc func(key string) ([]byte, error)

//将函数转换为接口的技巧：定义函数类型F，实现接口A的方法（一个方法），然后在方法中调用自己F

// Get 实现Getter接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	// Get将key-> 传给f实例，由f完成具体功能
	return f(key)
}

var (
	// 读写锁
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup create a new instance of Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get value for a key from cache
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	// 先查缓存
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 缓存未命中
	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
// 一个命名空间注册其对等点列表
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 可以选择从远端节点进行数据加载
	if g.peers != nil {
		// 先确定从哪个节点获取数据
		// 从一致性哈希环 上确定合适的节点
		if peer, ok := g.peers.PickPeer(key); ok {
			// 从远端节点获取数据
			if value, err = g.getFromPeer(peer, key); err == nil {
				return value, nil
			}
			log.Println("[GeeCache] Failed to get from peer", err)
		}
	}
	// 否则，从本地获取数据
	return g.getLocally(key)
}

// 加载进缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 本地加载数据进缓存
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	// 数据只读
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 从远程节点
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	// Get里面会通过http.Get请求数据
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
