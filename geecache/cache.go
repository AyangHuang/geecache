package geecache

import (
	"geecache/algorithm"
	"sync"
)

type cache struct {
	// 全局锁，粒度较大
	sync.Mutex
	// 接口的重要学习！！！！！！！自己学到的。通过项目学到的，不是通过geetoto指导的
	// 1. 不能指针，因为指针接口不能调用接口函数，但是普通接口可以调用接口函数。。。可以指针，然后解引用！！！
	// 2. 但是没必要指针接口，因为类型接收者是指针或者非指针，都可以把指针类型赋值给接口！！！特别重要！！！
	// 现在终于明白为什么普通类型接收者为什么要自动生成指针接收者了
	store      algorithm.Cache
	cacheBytes int
}

func newCache(algo string, maxBytes int) *cache {
	return &cache{
		cacheBytes: maxBytes,
		store:      algorithm.NewCache(algo, maxBytes, nil),
	}
}

func (c *cache) add(key string, value ByteView) {
	c.Lock()
	defer c.Unlock()
	if c.store != nil {
		c.store.Add(key, value)
	}
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.Lock()
	defer c.Unlock()
	if c.store != nil {
		if v, ok := c.store.Get(key); ok {
			value := v.(ByteView)
			return value, ok
		}
	}
	// 函数返回参数有初始化，这里不用了
	return
}
