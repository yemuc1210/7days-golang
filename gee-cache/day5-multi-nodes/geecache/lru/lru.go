package lru

import "container/list"

// Cache is a LRU cache. It is not safe for concurrent access.
// 并发不安全的Cache
type Cache struct {
	// 允许使用的最大内存
	maxBytes int64
	// 当前已经使用的内存  字节数
	nbytes int64
	// 双向链表  为淘汰策略服务
	ll *list.List
	// 字典定义   为查询-缓存命中服务
	cache map[string]*list.Element
	// 当条目被清除时，可选执行   被清除时使用的回调函数
	OnEvicted func(key string, value Value)
}

// 一个缓存项，包括k-v
type entry struct {
	key   string
	value Value
}

// 允许是任何类型，用接口存储数据
// 包含一个Len方法，返回值占用的内存大小
type Value interface {
	Len() int
}

// 构造器  实例化一个缓存Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 添加一个缓存项
func (c *Cache) Add(key string, value Value) {
	// 判断元素是否存在
	// 若存在
	if ele, ok := c.cache[key]; ok {
		// 采用LRU淘汰策略，所以，需要将刚访问的移动到队列前
		// 每次淘汰队尾的元素
		c.ll.MoveToFront(ele)
		// Element Value是interface类型，通过断言进行类型转换
		kv := ele.Value.(*entry)
		// 经过类型转换后，Cache占用的bytes发生变化    v的变化
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 新增
		// 将 k-v 通过Entry结构体包装
		// 在双向链表的队头插入
		// 返回一个 *Element  类型，即新插入的数据引用
		ele := c.ll.PushFront(&entry{key, value})
		// 将数组存入字典
		c.cache[key] = ele
		// 更新占用的内存数  k-v都存，所以变化量包括key+value
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 判断占用内存是否到达最大
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		// 根据淘汰策略，淘汰元素
		c.RemoveOldest()
	}
}

// Get look ups a key's value
// 根据键查询值
// 第一步：从字典中找到对应的双向链表节点；第二步：将该节点移动到队头
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// ele 类型为 *list.Element，也是双向链表的节点
		// 移动到链表头部
		c.ll.MoveToFront(ele)
		// 获取value返回
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最老的元素-lru淘汰鬼册
func (c *Cache) RemoveOldest() {
	// 返回链表尾部的元素
	ele := c.ll.Back()
	if ele != nil {
		// 移除尾部元素
		c.ll.Remove(ele) // 返回值是删除的Element元素的value、
		// 同时删除字典中的元素    目前为止，改缓存策略不是并发安全的
		kv := ele.Value.(*entry)
		// delete 用于删除map中元素
		delete(c.cache, kv.key)
		// 更新占用内存，此时是减去操作
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 判断是否需要调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len the number of cache entries
// 实现Value接口
func (c *Cache) Len() int {
	return c.ll.Len()
}
