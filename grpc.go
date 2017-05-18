package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/jgadling/pennant/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server is used to implement proto.PennantServer.
type server struct {
	fc *FlagCache
}

// GetFlagValue implements proto.PennantServer
func (s *server) GetFlagValue(ctx context.Context, in *pb.FlagRequest) (*pb.FlagReply, error) {
	flagCache := s.fc
	logger.Warning("Flag requested: ", in.Name)
	logger.Warning("Flag String Data: ", in.Strings)
	logger.Warning("Flag Number Data: ", in.Numbers)
	flag, err := flagCache.Get(in.Name)
	if err != nil {
		// The flag didn't exist in the cache, let's send a 404
		return &pb.FlagReply{Status: 404, Enabled: false}, nil
	}
	datas := make(map[string]interface{})
	for k, v := range in.Strings {
		datas[k] = v
	}
	for k, v := range in.Numbers {
		datas[k] = v
	}
	enabled := flag.GetValue(datas)
	return &pb.FlagReply{Status: 200, Enabled: enabled}, nil
}

func runGrpc(conf *Config, fc *FlagCache) {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.GrpcAddr, conf.GrpcPort))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterPennantServer(s, &server{fc: fc})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
