package singleflight

import "sync"

// 针对相同的Key，可能会发起多次请求。显然，这是需要避免的，因为http请求也是耗费资源的

// call is an in-flight or completed Do call
// call 类型
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group represents a class of work and forms a namespace in which
// units of work can be executed with duplicate suppression.
// Group 表示一组任务的命名空间，其中工作单元可以通过重复抑制来执行
// 管理不同key的请求（call）
type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
// 确保在一个时刻，对于相同的key，请求只会执行一次
// 无论Do被调用多少次， 函数fn只会被调用一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		// 键类型为string
		g.m = make(map[string]*call)
	}
	// 如果存在 同一个key 被请求过，那么这个任务是重复的
	if c, ok := g.m[key]; ok {
		// 发生重复调用，该调用者只需要等待原始调用者完成任务，因而不需要对m进行操作
		// 所以g.mu可以解锁
		g.mu.Unlock()
		// 重复的调用者等待原始调用者完成
		c.wg.Wait()
		// 完成并接受同一个结果    值是*call 需要得到这个请求调用的结果 -> val
		return c.val, c.err
	}
	// 否则，当前调用是”原始调用者“
	// 新建请求
	c := new(call)
	c.wg.Add(1)
	// 新建请求写入字典，使得之后重复的请求不必重复执行
	g.m[key] = c
	g.mu.Unlock()

	// 调用函数fn()
	// 将调用的执行结果写入call 结构
	c.val, c.err = fn()
	// 保证同一时间只执行一次，而不是一次执行之后就不执行了
	c.wg.Done()

	g.mu.Lock()
	// 将这个key对应的call删除
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
