package cache

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/einsier/go-cache/consistenthash"
	pb "github.com/einsier/go-cache/gocachepb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPool struct {
	pb.UnimplementedGroupCacheServer

	self        string
	mu          sync.Mutex
	peers       *consistenthash.Map
	grpcGetters map[string]*grpcGetter
}

func NewGRPCPool(self string) *GRPCPool {
	return &GRPCPool{
		self:        self,
		peers:       consistenthash.New(defaultReplicas, nil),
		grpcGetters: make(map[string]*grpcGetter),
	}
}

func (p *GRPCPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *GRPCPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers.Add(peers...)
	for _, peer := range peers {
		p.grpcGetters[peer] = &grpcGetter{addr: peer}
	}
}

func (p *GRPCPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.grpcGetters[peer], true
	}
	return nil, false
}

func (p *GRPCPool) Get(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	p.Log("Received Get request for group: %s, key: %s", in.Group, in.Key)

	group := GetGroup(in.Group)
	if group == nil {
		p.Log("no such group %v", in.Group)
		return nil, fmt.Errorf("no such group %v", in.Group)
	}
	// get value from group
	value, err := group.Get(in.Key)
	if err != nil {
		p.Log("get key %v error %v", in.Key, err)
		return nil, err
	}

	return &pb.Response{
		Value: value.ByteSlice(),
	}, nil
}

func (p *GRPCPool) Run() {
	lis, err := net.Listen("tcp", p.self)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGroupCacheServer(s, p)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

var _ PeerGetter = (*grpcGetter)(nil)

type grpcGetter struct {
	addr string
}

func (g *grpcGetter) Get(in *pb.Request, out *pb.Response) error {
	c, err := grpc.Dial(g.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer c.Close()

	client := pb.NewGroupCacheClient(c)
	res, err := client.Get(context.Background(), in)
	out.Value = res.Value
	return err
}
