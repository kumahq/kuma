package cmd

import (
	"fmt"

	"github.com/vishvananda/netns"

	. "github.com/kumahq/kuma/app/kuma-cni/pkg/log"
	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
)

func redirect(netNS string, pod *PodInfo) error {
	/* Save current network namespace */
	hostNs, err := netns.Get()
	if err != nil {
		Log.Errorf("failed getting host namespace: %v", err)

		return err
	}

	Log.Info("host namespace: ", hostNs)

	defer func() {
		if err := hostNs.Close(); err != nil {
			Log.Error("failed closing host namespace handle: ", err)

			return
		}

		Log.Info("closed host namespace handle: ", hostNs)
	}()

	targetNs, err := netns.GetFromPath(netNS)
	if err != nil {
		Log.Errorf("failed switching to desired namespace: %v", err)

		return err
	}

	/* Switch to the desired namespace */
	if err := netns.Set(targetNs); err != nil {
		Log.Errorf("failed switching to desired namespace: %v", err)

		return err
	}

	Log.Info("switched to desired namespace: ", targetNs)

	/* Don't forget to switch back to the host namespace */
	defer func() {
		if err := netns.Set(hostNs); err != nil {
			Log.Errorf("failed switching back to host namespace: %v", err)

			return
		}

		Log.Info("switched back to host namespace: ", hostNs)
	}()

	/* Verify we are in the desired namespace */
	currentNs, err := netns.Get()
	if err != nil {
		Log.Errorf("failed getting current namespace: %v", err)

		return err
	}

	Log.Info("current namespace: ", currentNs)

	if hostNs == currentNs {
		Log.Errorf("unable to switch from %v to %v", hostNs, currentNs)

		return fmt.Errorf("unable to switch from %v to %v", hostNs, currentNs)
	}

	tpRedirect, err := kubernetes.NewPodRedirectForPod(pod.pod)
	if err != nil {
		Log.Errorf("failed generate pod redirect: %v", err)

		return err
	}

	tp := transparentproxy.DefaultTransparentProxy()
	tpConfig := tpRedirect.AsTransparentProxyConfig()

	Log.Info("transparent proxy config: ", tpConfig)

	if _, err := tp.Setup(tpConfig); err != nil {
		Log.Errorf("failed to setup transparent proxy: %v", err)

		return err
	}

	return nil
}
