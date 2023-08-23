package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

var waitLog = dataplaneLog.WithName("wait")

func newWaitCmd() *cobra.Command {
	args := struct {
		url            string
		requestTimeout time.Duration
		timeout        time.Duration
		checkFrequency time.Duration
	}{
		url:            "http://localhost:9901/ready",
		requestTimeout: 500 * time.Millisecond,
		timeout:        60 * time.Second,
		checkFrequency: 1 * time.Second,
	}
	cmd := &cobra.Command{
		Use:   "wait",
		Short: "Waits for Dataplane to be ready",
		Long:  `Waits for Dataplane (Envoy) to be ready.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client := &http.Client{
				Timeout: args.requestTimeout,
			}
			ticker := time.NewTicker(args.checkFrequency)
			defer ticker.Stop()
			timeout := time.After(args.timeout)

			waitLog.Info("waiting for Dataplane to be ready for %s", args.timeout)
			for {
				select {
				case <-ticker.C:
					err := checkIfEnvoyReady(client, args.url)
					if err == nil {
						waitLog.Info("dataplane is ready!")
						return nil
					}
				case <-timeout:
					return fmt.Errorf("timeout occurred while waiting for Dataplane to be ready")
				}
			}
		},
	}

	cmd.PersistentFlags().DurationVar(&args.checkFrequency, "check-frequency", args.checkFrequency, `frequency of checking if the Dataplane is ready`)
	cmd.PersistentFlags().DurationVar(&args.timeout, "timeout", args.timeout, `timeout defines how long waits for the Dataplane`)
	cmd.PersistentFlags().DurationVar(&args.requestTimeout, "request-timeout", args.requestTimeout, `requestTimeout defines timeout for the request to the Dataplane`)
	cmd.PersistentFlags().StringVar(&args.url, "url", args.url, `url at which admin is exposed`)

	return cmd
}

func checkIfEnvoyReady(client *http.Client, url string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
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
