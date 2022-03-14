package geecache

import pb "geecache/geecachepb"

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
// 修改该接口，以适应protobuf的使用
type PeerGetter interface {
	// 参数改变，使用pb的数据类型
	Get(in *pb.Request, out *pb.Response) error
}
