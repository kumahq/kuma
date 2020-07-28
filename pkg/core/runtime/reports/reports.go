package reports

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

const (
	pingInterval = 3600
	pingHost     = "kong-hf.konghq.com"
	pingPort     = 61831
)

var log = core.Log.WithName("core").WithName("reports")

type reporter struct {
	resManager manager.ReadOnlyResourceManager

	sync.Mutex
	mutable   map[string]string
	immutable map[string]string
}

var _ component.Component = &reporter{}

func newReporter(config kuma_cp.Config, instanceID string, resourceManager manager.ReadOnlyResourceManager) *reporter {
	reporter := &reporter{
		resManager: resourceManager,
		mutable:    map[string]string{},
		immutable:  map[string]string{},
	}
	reporter.initImmutable(config, instanceID)
	return reporter
}

func Setup(rt core_runtime.Runtime) error {
	if !rt.Config().Reports.Enabled {
		return nil
	}
	return rt.Add(newReporter(rt.Config(), rt.GetInstanceId(), rt.ReadOnlyResourceManager()))
}

func (b *reporter) Start(stop <-chan struct{}) error {
	go func() {
		if err := b.dispatch(pingHost, pingPort, "start"); err != nil {
			log.V(2).Info("Failed sending usage info", "err", err)
		}
		ticker := time.NewTicker(time.Second * pingInterval)
		for {
			select {
			case <-ticker.C:
				if err := b.dispatch(pingHost, pingPort, "ping"); err != nil {
					log.V(2).Info("Failed sending usage info", "err", err)
				}
			case <-stop:
				return
			}
		}
	}()
	<-stop
	return nil
}

func (b *reporter) NeedLeaderElection() bool {
	return true
}

func (b *reporter) fetchDataplanes() (*mesh.DataplaneResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := b.resManager.List(context.Background(), &dataplanes); err != nil {
		return nil, errors.Wrap(err, "Could not fetch dataplanes")
	}

	return &dataplanes, nil
}

func (b *reporter) fetchMeshes() (*mesh.MeshResourceList, error) {
	meshes := mesh.MeshResourceList{}
	if err := b.resManager.List(context.Background(), &meshes); err != nil {
		return nil, errors.Wrap(err, "Could not fetch meshes")
	}

	return &meshes, nil
}

func (b *reporter) marshall() (string, error) {
	var builder strings.Builder

	_, err := fmt.Fprintf(&builder, "<14>")
	if err != nil {
		return "", err
	}

	for k, v := range b.immutable {
		_, err := fmt.Fprintf(&builder, "%s=%s;", k, v)
		if err != nil {
			return "", err
		}
	}

	for k, v := range b.mutable {
		_, err := fmt.Fprintf(&builder, "%s=%s;", k, v)
		if err != nil {
			return "", err
		}
	}

	return builder.String(), nil
}

// XXX this function retrieves all dataplanes and all meshes;
// ideally, the number of dataplanes and number of meshes
// should be pushed from the outside rather than pulled
func (b *reporter) updateEntitiesReport() error {
	dps, err := b.fetchDataplanes()
	if err != nil {
		return err
	}
	b.mutable["dps_total"] = strconv.Itoa(len(dps.Items))

	meshes, err := b.fetchMeshes()
	if err != nil {
		return err
	}
	b.mutable["meshes_total"] = strconv.Itoa(len(meshes.Items))
	return nil
}

func (b *reporter) dispatch(host string, port int, pingType string) error {
	if err := b.updateEntitiesReport(); err != nil {
		return err
	}
	b.mutable["signal"] = pingType
	pingData, err := b.marshall()
	if err != nil {
		return err
	}
	log.V(2).Info("dispatching usage statistics", "data", pingData)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(conn, pingData)
	if err != nil {
		return err
	}

	return nil
}

func (b *reporter) initImmutable(config kuma_cp.Config, instanceID string) {
	b.immutable["version"] = kuma_version.Build.Version
	b.immutable["unique_id"] = instanceID
	b.immutable["backend"] = config.Store.Type
	b.immutable["mode"] = config.Mode

	hostname, err := os.Hostname()
	if err == nil {
		b.immutable["hostname"] = hostname
	}
}
