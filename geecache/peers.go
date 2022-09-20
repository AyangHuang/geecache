package geecache

// 两层接口，其实都是为了解耦。太牛逼了。。。

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
