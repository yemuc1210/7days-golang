package geecache

// 为Cache添加注册和选择节点的功能，并实现HTTP客户端，与远程节点通信

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
// PeerPicker接口，根据传入的key选择相应的节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
// 用于从对应group中查找缓存值   客户端功能
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
