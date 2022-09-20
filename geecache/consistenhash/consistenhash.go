package consistenhash

import (
	"hash/crc32"
	"math/rand"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // hash函数
	replaces int            // 真实结点对应多少个虚拟结点
	keys     []int          // 虚拟结点，排序，从小到大，从0-2^32。但是为了排序。用int把
	hashMap  map[int]string // 根据虚拟结点找到真实结点，真实结点是ip
}

func NewMap(replaces int, hash Hash) *Map {
	m := &Map{
		hash:     hash,
		replaces: replaces,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(realNode ...string) {
	for _, v := range realNode {
		for i := 0; i < m.replaces; i++ {
			//这里不是定的，所以每个服务端的HttpPool的虚拟结点映射都不一样，所以可能出现循环转
			//hash := int(m.hash([]byte(randString(32) + v)))
			hash := int(m.hash([]byte(strconv.Itoa(i) + v)))
			m.hashMap[hash] = v
			m.keys = append(m.keys, hash)
		}
	}
	sort.Ints(m.keys)
}

// Get 根据key值选择合适的ip
func (m *Map) Get(key string) string {
	node := int(m.hash([]byte(key)))
	virtual := m.keys[0]
	for i := 0; i < len(m.keys)-1; i++ {
		if node > m.keys[i] {
			virtual = m.keys[i+1]
		}
	}
	return m.hashMap[virtual]
}

func randString(length int) string {
	str := "0123456789abcdefghiklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	res := make([]byte, length, length)
	for i := 0; i < length; i++ {
		//用1970到现在的纳秒数为种子，还是不行，变化太快了。种子是一样的
		//r := rand.New(rand.NewSource(time.Now().UnixNano()))
		//用固定种子把。
		num := rand.Intn(61)
		res[i] = str[num]
	}
	return string(res)
}
