package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const kumaDPProcessName = "kuma-dp"

// newDrainCmd sends SIGUSR2 to running kuma-dp processes so they start draining Envoy.
// It's used as the sidecar preStop hook, so the image doesn't need `killall`.
func newDrainCmd(procRoot string, signal func(pid int, sig syscall.Signal) error) *cobra.Command {
	return &cobra.Command{
		Use:   "drain",
		Short: "Signals running data plane proxies to start draining",
		Long:  `Sends SIGUSR2 to every other running kuma-dp process so it starts draining Envoy connections.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pids, err := findProcesses(procRoot, kumaDPProcessName, os.Getpid())
			if err != nil {
				return err
			}
			if len(pids) == 0 {
				return errors.Errorf("no %s process found", kumaDPProcessName)
			}
			for _, pid := range pids {
				if err := signal(pid, syscall.SIGUSR2); err != nil {
					return errors.Wrapf(err, "could not signal process %d", pid)
				}
				cmd.Printf("sent SIGUSR2 to %s (pid %d)\n", kumaDPProcessName, pid)
			}
			return nil
		},
	}
}

func findProcesses(procRoot string, name string, skipPid int) ([]int, error) {
	entries, err := os.ReadDir(procRoot)
	if err != nil {
		return nil, errors.Wrapf(err, "could not list %s", procRoot)
	}
	var pids []int
	for _, entry := range entries {
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid == skipPid {
			continue
		}
		comm, err := os.ReadFile(filepath.Join(procRoot, entry.Name(), "comm"))
		if err != nil {
			// the process could have exited in the meantime
			continue
		}
		if strings.TrimSpace(string(comm)) == name {
			pids = append(pids, pid)
		}
	}
	return pids, nil
}
