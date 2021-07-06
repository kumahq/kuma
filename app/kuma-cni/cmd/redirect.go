package cmd

import (
	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/app/kuma-cni/pkg/log"
	. "github.com/kumahq/kuma/app/kuma-cni/pkg/nsenter"
	"github.com/kumahq/kuma/pkg/transparentproxy"
	"github.com/kumahq/kuma/pkg/transparentproxy/kubernetes"
)

func redirect(netns string, pod *PodInfo) error {
	tp := transparentproxy.DefaultTransparentProxy()

	tpRedirect, err := kubernetes.NewPodRedirectForPod(pod.pod)
	if err != nil {
		return errors.Wrap(err, "failed to setup transparent proxy")
	}

	var output string
	err = NsEnter([]Namespace{
		{
			Path: netns,
			Type: NSTypeCGroup,
		},
	}, func() error {
		output, err = tp.Setup(tpRedirect.AsTransparentProxyConfig())
		if err != nil {
			return errors.Wrapf(err, "failed to setup transparent proxy:\n%s\n", output)
		}

		return nil
	})

	Log.Printf("%s\n", output)

	return err
}
