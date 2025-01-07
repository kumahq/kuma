package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/kumahq/kuma/test/server/grpc/api"
)

var grpcClientLog = grpcLog.WithName("client")

func newGRPCClientCmd() *cobra.Command {
	args := struct {
		address string
		unary   bool
	}{}
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Run GRPC Test Server",
		Long:  "Run GRPC Test Server.",
		Example: `
# Start a GRPC client that connects to localhost:8080, and
# sends "Request #${n_of_streams}.${n_of_requests}" every 1 seconds.
test-server grpc client address="localhost:8080" unary=true`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			streamCounter := 0
			for {
				streamCounter++
				if err := startSendingRequests(args.address, streamCounter, args.unary); err != nil {
					grpcClientLog.Error(err, "sending requests failed")
				}
				time.Sleep(1 * time.Second)
			}
		},
	}
	cmd.PersistentFlags().StringVar(&args.address, "address", "localhost:8080", "server address")
	cmd.PersistentFlags().BoolVar(&args.unary, "unary", false, "use unary RPC instead of streaming")
	return cmd
}

func startSendingRequests(address string, streamCounter int, unary bool) error {
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    10 * time.Second,
			Timeout: 10 * time.Second,
		}),
	)
	if err != nil {
		return errors.Wrap(err, "dial failed")
	}
	defer conn.Close()

	c := api.NewGreeterClient(conn)
	if unary {
		return startUnaryRequests(c, streamCounter)
	}
	return startStreamingRequests(c, streamCounter)
}

func startStreamingRequests(c api.GreeterClient, streamCounter int) error {
	for {
		clientStream, err := c.SayHellos(context.Background())
		if err != nil {
			return errors.Wrap(err, "SayHellos failed")
		}

		requestCounter := 0
		for {
			err := clientStream.Send(&api.HelloRequest{
				Name: fmt.Sprintf("Request #%d.%d", streamCounter, requestCounter),
			})
			if err != nil {
				return errors.Wrap(err, "Send failed")
			}

			resp, err := clientStream.Recv()
			if err != nil {
				return errors.Wrap(err, "Recv failed")
			}
			grpcClientLog.Info("received response", "msg", resp.GetMessage())

			time.Sleep(1 * time.Second)
			requestCounter++
		}
	}
}

func startUnaryRequests(c api.GreeterClient, streamCounter int) error {
	for {
		requestCounter := 0
		resp, err := c.SayHello(context.Background(), &api.HelloRequest{
			Name: fmt.Sprintf("Request #%d.%d", streamCounter, requestCounter),
		})
		if err != nil {
			return errors.Wrap(err, "Send failed")
		}
		grpcClientLog.Info("received response", "msg", resp.GetMessage())

		time.Sleep(1 * time.Second)
		requestCounter++
	}
}
