package main

import (
	"context"
	"fmt"
	"io"
	_ "math"
	"net"
	proto "sample-grpc/proto/github.com/example/path/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct{}

var prevMax int64
var totReq float32 = 0
var currAvg float32 = 0

func main() {
	listener, err := net.Listen("tcp", ":4040")
	if err != nil {
		panic(err)
	}

	svr := grpc.NewServer()
	proto.RegisterExamplesServer(svr, &server{})
	reflection.Register(svr)

	if e := svr.Serve(listener); e != nil {
		panic(e)
	}
}

func (s *server) Add(ctx context.Context, req *proto.AddRequest) (*proto.AddResponse, error) {
	a, b := req.GetA(), req.GetB()

	sum := a + b

	return &proto.AddResponse{Sum: sum}, nil
}

func (s *server) Maximum(stream proto.Examples_MaximumServer) error {
	for {
		req, err := stream.Recv()
		fmt.Println(req)
		if err == io.EOF {
			return stream.SendAndClose(&proto.MaximumResponse{Maximum: prevMax})
		}

		if err != nil {
			return err
		}

		num := req.GetA()
		if prevMax < num {
			prevMax = num
		}
	}
}

func (s *server) Multiply(req *proto.MultiplyRequest,stream proto.Examples_MultiplyServer) error {
	a := req.GetA()
	var i int64
	for i = 0; i < 10; i++ {
		table := fmt.Sprintf("%d * %d = %d",a,  i, a*i)
		stream.Send(&proto.MultiplyResponse{Table: table})
	}
	return nil
}

func (s *server) RunningAverage(stream proto.Examples_RunningAverageServer) error {
	for {
		req, err := stream.Recv()
		fmt.Println(req)
		if err != nil {
			return err
		}
		num := req.GetA()
		totReq++
		currAvg = (currAvg+num)/totReq
		fmt.Println("ans",currAvg)
		stream.Send(&proto.RunningAverageResponse{Average: currAvg})
	}
	return nil
}
