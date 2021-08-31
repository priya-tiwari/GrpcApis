package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	proto "sample-grpc/proto/github.com/example/path/gen"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	//"sample-grpc/proto"

	//proto "sample-grpc/proto"
	"strconv"
)

type server struct{}

func randomPoint(r *rand.Rand) *proto.MaximumRequest {
	a := (r.Int63n(180) - 90) * 1e7
	return &proto.MaximumRequest{A: a}
}

func main() {
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := proto.NewExamplesClient(conn)
	g := gin.Default()

	g.GET("/add/:a/:b", func(ctx *gin.Context) {
		a, err := strconv.ParseUint(ctx.Param("a"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter a"})
			return
		}

		b, err := strconv.ParseUint(ctx.Param("b"), 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter b"})
			return
		}

		req := &proto.AddRequest{A: int64(a), B: int64(b)}
		if response, err := client.Add(ctx, req); err == nil {
			ctx.JSON(http.StatusOK, gin.H{
				"result": fmt.Sprint(response.Sum),
			})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	})

	g.GET("/multiply/:a", func(c *gin.Context) {
		num , err := strconv.ParseInt(c.Param("a"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter b"})
			return
		}
		stream , err := client.Multiply(context.Background(), &proto.MultiplyRequest{A: num})
		if err != nil {
			panic(err)
		}
		for {
			feature, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}
			log.Println(feature)
		}
	})

	g.GET("/max/:a", func(c *gin.Context) {
		num := make([]int64, 0)
		num = append(num,2,46,6,1,98,45,23,44,566,22)
		stream, err := client.Maximum(context.Background())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
			return
		}
		for _ , point := range num {
			fmt.Println("Sending req:")
			stream.Send(&proto.MaximumRequest{A:point});
			time.Sleep(10*time.Millisecond)
		}

		maxi, err := stream.CloseAndRecv()
		if err != nil {
			panic(err)
		}
		fmt.Println("eidnx",maxi.GetMaximum())
		c.JSON(http.StatusOK, gin.H{
			"Maximum": fmt.Sprint(maxi),
		})
	})

	g.GET("/runavg/:a", func(c *gin.Context) {
		num := make([]float32, 0)
		num = append(num,2,46,6,1,98,45,23,44,566,22)
		stream, err := client.RunningAverage(context.Background())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parameter"})
			return
		}
		for _ , point := range num {
			stream.Send(&proto.RunningAverageRequest{A:point});
			feature, err := stream.Recv()

			if err != nil {
				panic(err)
			}
			log.Println("Running Average after",point," is ",feature)
		}
	})

	if err := g.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
