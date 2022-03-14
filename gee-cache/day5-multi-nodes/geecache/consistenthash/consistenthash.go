package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 一致性哈希：缓存从单节点走向分布式节点的重要环节

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed keys
type Map struct {
	// 定义函数类型Hash，依赖注入，允许替换成自定义的Hash函数，默认是hash/crc32/ChecksumIEEE算法
	hash Hash
	// 虚拟节点倍数，解决数据倾斜的问题
	replicas int
	// 哈希环
	keys []int // Sorted
	// 虚拟节点和真实节点之间的映射表
	// 键是虚拟节点的哈希值，值是真实节点的名称
	hashMap map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	// 允许用户自定义Hash函数，否则使用默认的hash函数
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
// 允许传入0或多个 真实节点名称，keys 类型是string
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对于每个真实节点key，创建虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点的名称是i+key -  加编号    真实节点是不带编号的节点
			// 根据虚拟节点的名称计算哈希值
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 虚拟节点上哈希环  存的是哈希值
			m.keys = append(m.keys, hash)
			// 更新虚拟节点和真实节点的映射
			m.hashMap[hash] = key
		}
	}
	// 哈希环上的哈希值排序  【每次Add都需要排序，复杂度略高啊】
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
// 获得哈希环中，与key离得最近的节点
// 根据一致性哈希算法，从计算的hash值顺时针遍历，遇到的第一个节点为所求节点
func (m *Map) Get(key string) string {
	// 哈希环上没有元素，直接返回
	if len(m.keys) == 0 {
		return ""
	}
	// 计算传入key的哈希
	hash := int(m.hash([]byte(key)))
	// Binary search for appropriate replica.
	// 二分查找，定义true规则，返回满足true的最小idx  查找范围[0,n)
	idx := sort.Search(len(m.keys), func(i int) bool {
		// 定义true规则
		return m.keys[i] >= hash
	})
	// 从哈希环上取对应的哈希值，然后查虚拟-真实节点映射表，获得真实的节点名称
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
