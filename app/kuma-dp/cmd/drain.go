package cmd

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func newDrainCmd() *cobra.Command {
	args := struct {
		port    int
		timeout time.Duration
	}{}
	cmd := &cobra.Command{
		Use:   "drain",
		Short: "Tells Envoy to begin draining forever",
		Long:  `Tells Envoy to begin draining forever`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return triggerDrain(args.port)
		},
	}

	cmd.PersistentFlags().IntVar(&args.port, "port", 9901, `kuma-dp admin server port`)

	return cmd
}

func triggerDrain(port int) error {
	url := fmt.Sprintf("http://localhost:%d/drain", port)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
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
