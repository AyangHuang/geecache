package algorithm

import (
	"container/list"
)

// entry 是 list element 里面的value
// list 一定要记住key，因为删除缓存是从链表反找key，所以一定要存储key
type entry struct {
	key   string
	value Value
}

type LRUCache struct {
	l         *list.List
	m         map[string]*list.Element      //只记录list元素的地址，实际缓存是存在链表中
	maxBytes  int                           // 允许使用的最大内存
	nowBytes  int                           // 当前已经使用的内存
	onEvicted func(key string, value Value) // 记录被移除时的回调函数，默认为nil
}

func newLRU(maxBytes int, onEvicted func(key string, value Value)) *LRUCache {
	return &LRUCache{
		l:         list.New(),
		m:         make(map[string]*list.Element),
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
	}
}

// Get
// 1. 有缓存，更新到链表最前头，并返回
// 2. 无缓存，直接返回
func (this *LRUCache) Get(key string) (Value, bool) {
	if element, ok := this.m[key]; ok {
		//用到了就要挪到最前面
		this.l.MoveToFront(element)
		v := element.Value.(*entry)
		return v.value, true
	}
	return nil, false
}

// Add
// 1. 有缓存，提到最前面去，并跟新（需判断跟新会否因为value过大而溢出）
// 2. 没有缓存，加入缓存
//	（1）缓存已满，提出链表尾部的缓存，记得map也要删除
//  （2）缓存未满，加到最前面去
func (this *LRUCache) Add(key string, value Value) {
	if value.Len() > this.maxBytes {
		return
	}
	//缓存已经存在
	if element, ok := this.m[key]; ok {
		this.l.MoveToFront(element)
		entry := element.Value.(*entry)
		// 判断跟新会否因为value过大而溢出
		if value.Len() > entry.value.Len() && this.maxBytes < this.nowBytes-entry.value.Len()+value.Len() {
			this.flush(entry.value.Len() - value.Len())
		}
		// 更新value
		entry.value = value
		this.nowBytes += value.Len()
	} else {
		//缓存不足，先释放足够的空间
		if this.maxBytes-this.nowBytes < value.Len() {
			this.flush(value.Len() - (this.maxBytes - this.nowBytes))
		}
		//新加入缓存 头插法
		// 存地址，类型断言的时候才不会copy一份
		e := this.l.PushFront(
			&entry{
				key:   key,
				value: value,
			})
		this.m[key] = e
		this.nowBytes += value.Len()
	}
}

func (this *LRUCache) flush(need int) {
	for need > 0 {
		delLen := this.removeOldest()
		need -= delLen
	}
}

func (this *LRUCache) removeOldest() int {
	delE := this.l.Back()
	delV := this.l.Remove(delE).(*entry)
	delete(this.m, delV.key)
	len := delV.value.Len()
	this.nowBytes -= len
	if this.onEvicted != nil {
		this.onEvicted(delV.key, delV.value)
	}
	return len
}

func (this *LRUCache) Len() int {
	return this.l.Len()
}
