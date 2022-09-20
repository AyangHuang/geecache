package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

var (
	// 读写锁 并发map，可以用sync.Map替代
	rw     sync.RWMutex
	groups = make(map[string]*Group)
)

type Getter interface {
	// Get 从本地数据源获取数据（例如mysql)
	Get(key string) ([]byte, error)
}

// GetterFunc  是一个实现了接口的函数类型，简称为接口型函数。
// 作用：既能够将普通的函数类型（需类型转换）作为参数，
// 也可以将结构体作为参数，使用更为灵活，可读性也更好，这就是接口型函数的价值。
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name   string
	getter Getter // 从数据源取出缓存没有的数据
	//下面这俩玩意都是接口（cache封装了一层，底层还是接口），面向接口编程，依赖倒转原则
	//就是说我仅仅只需要用到接口所拥有的函数。
	// 例如Cache接口我只要 Add() Get()函数
	//   peerPicker 我只需要用到picker函数，获得结点。然后 结点接口又只需要Get()函数，从结点获取数据。这里是两个接口，两层解耦！！！
	mainCache *cache     // 缓存
	peers     PeerPicker // 获取结点接口，两层解耦其实，下面全部用的是接口函数，没有用到其他函数
	loads     *singleflight.Group
}

func NewGroup(name string, algo string, getter Getter, maxBytes int) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: newCache(algo, maxBytes),
		loads:     &singleflight.Group{},
	}
	rw.Lock()
	groups[name] = group
	rw.Unlock()
	return group
}

func GetGroup(name string) *Group {
	rw.RLock()
	g := groups[name]
	rw.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		return
		//panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

//load 1.先判断是否应该从远程结点获取，
// 2. 1. 是，从远程结点获取
// 2. 2. 否，从本地数据源获取，获取并添加到缓存值
func (g *Group) load(key string) (value ByteView, err error) {
	v, err := g.loads.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})
	if err == nil {
		return v.(ByteView), nil
	}
	return
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 从本地数据源获取数据，并添加到缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 从远程结点获取缓存
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
