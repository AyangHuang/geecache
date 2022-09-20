package algorithm

type Value interface {
	Len() int
}
type Cache interface {
	Add(key string, value Value)
	Get(key string) (Value, bool)
}

// NewCache 工厂模式
func NewCache(algo string, maxBytes int, onEvicted func(key string, value Value)) Cache {
	var cache Cache
	switch algo {
	case "LRU":
		// 返回指针类型，但是因为指针类型的变量是接收者，所以可以用接口接受
		cache = newLRU(maxBytes, onEvicted)
	case "LFU":
		cache = newLFU(maxBytes, onEvicted)
	default:
		cache = newLRU(maxBytes, onEvicted)
	}
	return cache
}
