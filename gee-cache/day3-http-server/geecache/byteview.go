package geecache

// A ByteView holds an immutable view of bytes.
// 缓存值的只读视图
type ByteView struct {
	// 存储真实的缓存值
	b []byte
}

// Len returns the view's length
func (v ByteView) Len() int {
	return len(v.b)
}

// ByteView是只读的，使用该方法返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 将存储的[]byte数据以string类型返回
func (v ByteView) String() string {
	return string(v.b)
}

// 返回一个拷贝
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	//func copy(dst []Type, src []Type) int
	copy(c, b)
	return c
}
