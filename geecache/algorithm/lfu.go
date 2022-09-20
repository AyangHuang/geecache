package algorithm

import "container/list"

// eValue 是 list element 里面的value
// list 一定要记住key，因为删除缓存是从链表反找key，所以一定要存储key
type eValue struct {
	key   string
	value Value
	fre   int //频率
}
type LFUCache struct {
	mFre      map[int]*list.List
	mKey      map[string]*list.Element
	nowBytes  int
	maxBytes  int
	minFre    int // 维护最小频率
	onEvicted func(key string, value Value)
}

func newLFU(maxBytes int, onEvicted func(key string, value Value)) *LFUCache {
	return &LFUCache{
		mFre:      make(map[int]*list.List),
		mKey:      make(map[string]*list.Element),
		maxBytes:  maxBytes,
		onEvicted: onEvicted,
	}
}

// Get
// 1. mKye 直接找到，fre++，根据mFre查找插入（记得判断有无fre链表），且记得删除原来的。判断是否需要修改最低频率
// 2. 找不到
// 注意：因为element有list指针（表明属于哪个链表），且不能更改，所以把一个element移动到另一个list，不能复用，只能浅拷贝
//      同时，list只支持增加的只能是any，然后自动转换为element，并且返回，好在有返回element，也必须返回。。。
func (this *LFUCache) Get(key string) (Value, bool) {
	if element, ok := this.mKey[key]; ok {
		return this.addFre(key, element), true
	} else {
		return nil, false
	}
}

// Add
// 1. mKey （1）找得到，修改并改fre，步骤和get一样
//         （2) 没找到，添加两个哈希表
//					（1）哈希表未满，直接添加
//					（2）哈希表已满，去key去掉最少的fre的链表的最后一个，然后添加
func (this *LFUCache) Add(key string, value Value) {
	//lj lc ，居然让capacity为0...
	//if this.capacity == 0 {
	//	return
	//}
	element, ok := this.mKey[key]
	if ok {
		evalue, _ := element.Value.(eValue)
		// 判断跟新会否因为value过大而溢出
		if value.Len() > evalue.value.Len() && this.maxBytes < this.nowBytes-evalue.value.Len()+value.Len() {
			this.flush(evalue.value.Len() - value.Len())
		}
		//下面实际没跟新，断言返回全新的变量的原因
		//evalue, _ := element.Value.(eValue)
		//evalue.value = value
		//跟新fre
		//正确跟新value如下
		element.Value = eValue{value: value, key: key, fre: evalue.fre}
		this.addFre(key, element)
	} else {
		//缓存不足，先释放足够的空间
		if this.maxBytes-this.nowBytes < value.Len() {
			this.flush(value.Len() - (this.maxBytes - this.nowBytes))
		}
		this.addTwoMap(key, eValue{
			key:   key,
			value: value,
			fre:   1,
		})
		//不管如何直接改最低频率为1即可
		this.minFre = 1
	}
}

// addFre 跟新fre，同时返回value
// 1. 删除原来的双向链表的元素
// 2. 判断这个元素是否是最低频率，且最低频率链表只有它，那么需要需要更改最低频率
// 3. 向fre++双向链表加结点，同时修改mKey指向的element
func (this *LFUCache) addFre(key string, element *list.Element) Value {
	//断言应该是复制返回，但是这里无所谓
	evalue, _ := element.Value.(eValue)
	oldFre := evalue.fre
	evalue.fre++
	listFre, _ := this.mFre[oldFre]
	//删除原来的双向链表的元素
	//判断这个元素是否是最低频率，且最低频率链表只有它，那么需要更改最低频率
	if this.minFre == oldFre && listFre.Len() == 1 {
		this.minFre++
	}
	listFre.Remove(element)
	this.nowBytes -= evalue.value.Len()
	//增加或更新到两个map中
	this.addTwoMap(key, evalue)
	return evalue.value
}

// addTwoMap 加入或者更新element到两个map中
// 0. 判断缓存是否已满
// 1. 增加到双向链表
// 2. 增加（或更改）mKep 映射到新的element
func (this *LFUCache) addTwoMap(key string, evalue eValue) {
	//如果没有双向链表就创建,并且加进mFre
	listFre, ok := this.mFre[evalue.fre]
	if !ok {
		listFre = list.New()
		this.mFre[evalue.fre] = listFre
	}
	//向双向链表加结点，浅拷贝一份
	newE := listFre.PushFront(eValue{
		key:   evalue.key,
		value: evalue.value,
		fre:   evalue.fre,
	})
	this.nowBytes += evalue.value.Len()
	//跟换或增加mKey成新的element
	this.mKey[key] = newE
}

// 判断缓存是否已满
func (this *LFUCache) flush(need int) {
	for need > 0 {
		delLen := this.removeOldest()
		need -= delLen
	}
}

// 判断缓存已满且删除一个，这里没修改fre哈，没法找上一个fre
// 设计本来也没能找到上一个fre，因为只是在添加缓存的时候需要删除缓存
// 但实际添加后最小的fre 就是1
func (this *LFUCache) removeOldest() int {
	minFre, _ := this.mFre[this.minFre]
	element := minFre.Back()
	minFre.Remove(element)
	evalue, _ := element.Value.(eValue)
	delete(this.mKey, evalue.key)
	length := evalue.value.Len()
	this.nowBytes -= length
	if this.onEvicted != nil {
		this.onEvicted(evalue.key, evalue.value)
	}
	return length
}

func (this *LFUCache) Len() int {
	return len(this.mKey)
}
