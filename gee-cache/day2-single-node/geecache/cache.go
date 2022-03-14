package geecache

import (
	"geecache/lru"
	"sync"
)

// 为cache添加并发特性
// cache封装在Group结构中，因而首字母小写
type cache struct {
	// 使用锁添加并发特性
	mu sync.Mutex
	// lru缓存
	lru        *lru.Cache
	cacheBytes int64
}

// 对lru的add方法进行封装
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 根据需要实例化lru缓存
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	// 封装Add方法，并且加锁，支持并发
	c.lru.Add(key, value)
}

// 对lru cache的Get方法进行封装
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		// 空直接返回
		return
	}
	// 成功Get
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	// 否则，没有成功Get，返回空视图
	return
}
