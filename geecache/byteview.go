package geecache

// ByteView 缓存entry 的 value
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

// ByteSlice 返回的是复制的值，防止缓存被更改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 转为string 其实是默认不能更改了（当然可以通过unsafe包进行更改）
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
