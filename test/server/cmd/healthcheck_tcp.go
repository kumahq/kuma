package cmd

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"
)

func newHealthCheckTCP() *cobra.Command {
	args := struct {
		send     string
		recv     string
		port     uint32
		request  string
		response string
	}{}
	cmd := &cobra.Command{
		Use:   "tcp",
		Short: "Run Test Server for TCP Health Check test",
		Long:  `Run Test Server for TCP Health Check test.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ln, err := net.Listen("tcp", fmt.Sprintf(":%d", args.port))
			if err != nil {
				return err
			}
			defer func() {
				_ = ln.Close()
			}()

			for {
				conn, err := ln.Accept()
				if err != nil {
					return err
				}
				go func() {
					healthCheckLog.Info("accept new connection")

					s := ""
					hcIdx := -1
					buf := make([]byte, 1024)

					for {
						length, err := conn.Read(buf)
						if err != nil {
							healthCheckLog.Error(err, "error read from connection")
							break
						}

						s += string(buf[:length])
						healthCheckLog.V(1).Info("receive from connection", "buffer", s)

						if idx := strings.LastIndex(s, args.request); idx != -1 {
							healthCheckLog.Info("receive new request")
							if _, err = conn.Write([]byte(args.response)); err != nil {
								healthCheckLog.Error(err, "error write response to connection")
								return
							}
							healthCheckLog.Info("send response, close connection")
							if err := conn.Close(); err != nil {
								healthCheckLog.Error(err, "error during closing connection")
							}
							break
						}

						if idx := strings.LastIndex(s, args.send); hcIdx != idx && idx != -1 {
							healthCheckLog.Info("receive new health check request")
							if _, err = conn.Write([]byte(args.recv)); err != nil {
								healthCheckLog.Error(err, "error write health check response to connection")
								return
							}
							hcIdx = idx
						}
					}
				}()
			}
		},
	}
	cmd.PersistentFlags().Uint32Var(&args.port, "port", 10011, "port server is listening on")
	cmd.PersistentFlags().StringVar(&args.send, "send", "foo", "line that health checker sends")
	cmd.PersistentFlags().StringVar(&args.recv, "recv", "bar", "line that health checker expects to receive")
	cmd.PersistentFlags().StringVar(&args.request, "request", "request", "line that server can reply on")
	cmd.PersistentFlags().StringVar(&args.response, "response", "response", "server response")
	return cmd
}
