package cmd

import (
	"context"
	"fmt"
	"time"

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
	}{}
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Run GRPC Test Server",
		Long:  "Run GRPC Test Server.",
		Example: `
# Start a GRPC client that connects to localhost:8080, opens stream and 
# sends "Request #${n_of_streams}.${n_of_requests}" every 2 seconds.
test-server grpc client address="localhost:8080"`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conn, err := grpc.Dial(args.address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:    10 * time.Second,
					Timeout: 10 * time.Second,
				}),
			)
			if err != nil {
				grpcClientLog.Error(err, "dial failed")
				return err
			}
			defer conn.Close()

			c := api.NewGreeterClient(conn)

			streamCounter := 0
			for {
				clientStream, err := c.SayHellos(context.Background())
				if err != nil {
					grpcClientLog.Error(err, "SayHellos failed")
					return err
				}

				requestCounter := 0
				for {
					err := clientStream.Send(&api.HelloRequest{
						Name: fmt.Sprintf("Request #%d.%d", streamCounter, requestCounter),
					})
					if err != nil {
						grpcClientLog.Error(err, "Send failed")
						break
					}

					resp, err := clientStream.Recv()
					if err != nil {
						grpcClientLog.Error(err, "Recv failed")
						break
					}
					grpcClientLog.Info("received response", "msg", resp.GetMessage())

					time.Sleep(2 * time.Second)
					requestCounter++
				}

				streamCounter++
			}
		},
	}
	cmd.PersistentFlags().StringVar(&args.address, "address", "localhost:8080", "server address")
	return cmd
}
