package main

import (
	"context"
	"flag"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"

	"github.com/sirupsen/logrus"
	"github.com/yxyc/grpclb/balancer"
	pb "github.com/yxyc/grpclb/cmd/helloworld"
	grpclb "github.com/yxyc/grpclb/etcdv3"
)

var (
	svc = flag.String("service", "hello_service", "service name")
	reg = flag.String("reg", "http://localhost:2379", "register etcd address")
)

func main() {
	flag.Parse()
	r := grpclb.NewResolver([]string{*reg}, *svc)
	resolver.Register(r)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	// The gRPC client library will use the specified scheme to pick the right resolver plugin and pass it the fully qualified name string.
	conn, err := grpc.DialContext(ctx, r.Scheme()+"://authority/"+*svc, grpc.WithInsecure(), grpc.WithBalancerName(balancer.ConsistentHash), grpc.WithBlock())
	cancel()
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		client := pb.NewGreeterClient(conn)
		ctx := context.WithValue(context.Background(), balancer.DefaultConsistentHashKey, "test")
		resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			logrus.Infof("%v: Reply is %s\n", t, resp.Message)
		}
	}
}
