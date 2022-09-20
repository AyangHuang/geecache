package singleflight

import (
	"sync"
	"time"
)

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	sync.Mutex
	sync.Once
	m map[string]*call
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 懒加载且只调用一次，即单例
	g.Once.Do(func() {
		g.m = make(map[string]*call)
	})

	g.Lock()

	if c, ok := g.m[key]; ok {
		g.Unlock()  // 解锁，让其他并发的请求可以进来
		c.wg.Wait() // 等待get获取
		return c.val, c.err
	}

	c := &call{}
	g.m[key] = c
	c.wg.Add(1) // 让等待
	g.Unlock()  // 解锁让其他并发的进去等待

	c.val, c.err = fn()
	time.Sleep(10 * time.Second)
	c.wg.Done() // 让等待的其他并发获取数据

	g.Lock()
	delete(g.m, key) //删除key，并发结束
	g.Unlock()

	return c.val, c.err
}
