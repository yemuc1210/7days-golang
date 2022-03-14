package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 利用sync.Mutex实现并发控制

// 缓存命名空间，用于加载相关数据
// 当缓存不存在时，调用回调函数获取源数据
// 负责与用户的交互，因为是public，首字母大写
type Group struct {
	// 唯一的名称 name   缓存命名空间
	name string
	// 如何从数据源获取数据，是用户考虑的
	getter Getter
	// 缓存-数据域  另一个文件中实现
	// 实现并发读写的缓存
	mainCache cache
}

// Getter 加载key对应的数据到缓存中
// 具体如何获得数据，由用户决定
type Getter interface {
	// 定义Get的函数签名
	Get(key string) ([]byte, error)
}

// 定义函数类型 GetterFunc
// GetterFunc 使用一个函数实现Getter
// GetterFunc 是接口型函数
// 接口型函数：适用于接口只含有一个方法的情况
type GetterFunc func(key string) ([]byte, error)

//将函数转换为接口的技巧：定义函数类型F，实现接口A的方法（一个方法），然后在方法中调用自己F

// 实现Getter接口
func (f GetterFunc) Get(key string) ([]byte, error) {
	// Get将key-> 传给f实例，由f完成具体功能
	return f(key)
}

// 定义全局变量
var (
	// 读写锁
	mu sync.RWMutex
	// 命名空间groups
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
// 返回命名空间 or nil
func GetGroup(name string) *Group {
	mu.RLock()
	// map读取操作，要么读到值，要么为值类型的nil/零值
	g := groups[name] // 读操作+锁
	mu.RUnlock()
	return g
}

// Get value for a key from cache
// 读缓存
func (g *Group) Get(key string) (ByteView, error) {
	// ByteView 只读数据结构，表示缓存值
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 根据key读缓存
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	// 缓存未命中，需要将数据加载至缓存
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 根据需要判断从哪里获取key
	// 一种是需要从远程节点获取
	// 一种是本地获取，直接调用回调函数，获取值并加入缓存
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	// 通过回调函数，加载数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	//
	value := ByteView{b: cloneBytes(bytes)}
	// 填充缓存，即将加载结构写入缓存当中
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
