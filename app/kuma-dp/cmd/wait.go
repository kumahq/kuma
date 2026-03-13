package cmd

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

var waitLog = dataplaneLog.WithName("wait")

func newWaitCmd() *cobra.Command {
	args := struct {
		url            string
		unixSocket     string
		requestTimeout time.Duration
		timeout        time.Duration
		checkFrequency time.Duration
	}{}
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Waits for data plane proxy to be ready",
		Long:  `Waits for data plane proxy (Envoy) to be ready.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			var client *http.Client
			targetURL := args.url

			if args.unixSocket != "" {
				dialer := &net.Dialer{
					Timeout: args.requestTimeout,
				}
				client = &http.Client{
					Timeout: args.requestTimeout,
					Transport: &http.Transport{
						DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
							return dialer.DialContext(ctx, "unix", args.unixSocket)
						},
					},
				}
				targetURL = "http://localhost/ready"
			} else {
				client = &http.Client{
					Timeout: args.requestTimeout,
				}
			}

			ticker := time.NewTicker(args.checkFrequency)
			defer ticker.Stop()
			timeout := time.After(args.timeout)

			waitLog.Info("waiting for data plane proxy to be ready", "timeout", args.timeout)
			for {
				select {
				case <-ticker.C:
					if err := checkIfEnvoyReady(client, targetURL, args.unixSocket != ""); err != nil {
						waitLog.Info("data plane proxy is not ready", "err", err)
					} else {
						waitLog.Info("data plane is ready")
						return nil
					}
				case <-timeout:
					return fmt.Errorf("timeout occurred while waiting for data plane proxy to be ready")
				}
			}
		},
	}

	cmd.PersistentFlags().DurationVar(&args.checkFrequency, "check-frequency", time.Second, `frequency of checking if the data plane proxy is ready`)
	cmd.PersistentFlags().DurationVar(&args.timeout, "timeout", 180*time.Second, `timeout defines how long waits for the data plane proxy`)
	cmd.PersistentFlags().DurationVar(&args.requestTimeout, "request-timeout", 500*time.Millisecond, `requestTimeout defines timeout for the request to the data plane proxy`)
	cmd.PersistentFlags().StringVar(&args.url, "url", "http://localhost:9901/ready", `url at which admin is exposed`)
	cmd.PersistentFlags().StringVar(&args.unixSocket, "unix-socket", "", `path to Unix domain socket for admin API (overrides --url)`)

	return cmd
}

func checkIfEnvoyReady(client *http.Client, rawURL string, isUDS bool) error {
	if !isUDS {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL %q: %w", rawURL, err)
		}
		if host := parsed.Hostname(); host != "localhost" && host != "127.0.0.1" && host != "::1" {
			return fmt.Errorf("URL host must be a loopback address, got %q", host)
		}
	}
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := client.Do(req) // #nosec G704 -- URL validated to loopback above
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP status code %v", resp.StatusCode)
	}
	return nil
}
