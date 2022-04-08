package onionRPC

//
//import (
//	"errors"
//
//	"github.com/DistributedClocks/tracing"
//)
//
//type ServerStart struct {
//	NodeId uint8
//}
//
//type ServerJoining struct {
//	NodeId uint8
//}
//
//type NextServerJoining struct {
//	NextServerId uint8
//}
//
//type NewJoinedSuccessor struct {
//	NextServerId uint8
//}
//
//type ServerJoined struct {
//	NodeId uint8
//}
//
//type ServerFailRecvd struct {
//	FailedServerId uint8
//}
//
//type NewFailoverSuccessor struct {
//	NewNextServerId uint8
//}
//
//type NewFailoverPredecessor struct {
//	NewPrevServerId uint8
//}
//
//type ServerFailHandled struct {
//	FailedServerId uint8
//}
//
//type PutRecvd struct {
//	ClientId string
//	OpId     uint32
//	Key      string
//	Value    string
//}
//
//type PutOrdered struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//	Value    string
//}
//
//type PutFwd struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//	Value    string
//}
//
//type PutFwdRecvd struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//	Value    string
//}
//
//type PutResult struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//	Value    string
//}
//
//type GetRecvd struct {
//	ClientId string
//	OpId     uint32
//	Key      string
//}
//
//type GetOrdered struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//}
//
//type GetResult struct {
//	ClientId string
//	OpId     uint32
//	GId      uint64
//	Key      string
//	Value    string
//}
//
//type NodeConfig struct {
//	NodeId          uint8
//	CoordAddr         string
//	RelayAddr        string
//	ServerServerAddr  string
//	ServerListenAddr  string
//	ClientListenAddr  string
//	TracingServerAddr string
//	Secret            []byte
//	TracingIdentity   string
//}
//
//type Node struct {
//	// Node state may go here
//}
//
//func NewServer() *Node {
//	return &Node{}
//}
//
//func (s *Node) Start(serverId uint8, coordAddr string, serverAddr string, serverServerAddr string, serverListenAddr string, clientListenAddr string, strace *tracing.Tracer) error {
//	return errors.New("not implemented")
//}
