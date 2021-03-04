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

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"

	"github.com/pkg/errors"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	kuma_version "github.com/kumahq/kuma/pkg/version"
)

const (
	pingInterval = 3600
	pingHost     = "kong-hf.konghq.com"
	pingPort     = 61831
)

var (
	log = core.Log.WithName("core").WithName("reports")
)

/*
  - buffer initialized upon Init call
  - append adds more keys onto it
*/

type reportsBuffer struct {
	sync.Mutex
	mutable   map[string]string
	immutable map[string]string
}

func fetchDataplanes(rt core_runtime.Runtime) (*mesh.DataplaneResourceList, error) {
	dataplanes := mesh.DataplaneResourceList{}
	if err := rt.ReadOnlyResourceManager().List(context.Background(), &dataplanes); err != nil {
		return nil, errors.Wrap(err, "Could not fetch dataplanes")
	}

	return &dataplanes, nil
}

func fetchMeshes(rt core_runtime.Runtime) (*mesh.MeshResourceList, error) {
	meshes := mesh.MeshResourceList{}
	if err := rt.ReadOnlyResourceManager().List(context.Background(), &meshes); err != nil {
		return nil, errors.Wrap(err, "Could not fetch meshes")
	}

	return &meshes, nil
}

func fetchZones(rt core_runtime.Runtime) (*system.ZoneResourceList, error) {
	zones := system.ZoneResourceList{}
	if err := rt.ReadOnlyResourceManager().List(context.Background(), &zones); err != nil {
		return nil, errors.Wrap(err, "Could not fetch zones")
	}

	return &zones, nil
}

func (b *reportsBuffer) marshall() (string, error) {
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
func (b *reportsBuffer) updateEntitiesReport(rt core_runtime.Runtime) error {
	dps, err := fetchDataplanes(rt)
	if err != nil {
		return err
	}
	b.mutable["dps_total"] = strconv.Itoa(len(dps.Items))

	meshes, err := fetchMeshes(rt)
	if err != nil {
		return err
	}
	b.mutable["meshes_total"] = strconv.Itoa(len(meshes.Items))

	switch rt.Config().Mode {
	case config_core.Standalone:
		b.mutable["zones_total"] = strconv.Itoa(1)
	case config_core.Global:
		zones, err := fetchZones(rt)
		if err != nil {
			return err
		}
		b.mutable["zones_total"] = strconv.Itoa(len(zones.Items))
	}
	return nil
}

func (b *reportsBuffer) dispatch(rt core_runtime.Runtime, host string, port int, pingType string) error {
	if err := b.updateEntitiesReport(rt); err != nil {
		return err
	}
	b.mutable["signal"] = pingType
	b.mutable["cluster_id"] = rt.GetClusterId()
	pingData, err := b.marshall()
	if err != nil {
		return err
	}

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

// Append information to the mutable portion of the reports buffer
func (b *reportsBuffer) Append(info map[string]string) {
	b.Lock()
	defer b.Unlock()

	for key, value := range info {
		b.mutable[key] = value
	}
}

func (b *reportsBuffer) initImmutable(rt core_runtime.Runtime) {
	b.immutable["version"] = kuma_version.Build.Version
	b.immutable["product"] = kuma_version.Product
	b.immutable["unique_id"] = rt.GetInstanceId()
	b.immutable["backend"] = rt.Config().Store.Type
	b.immutable["mode"] = rt.Config().Mode

	hostname, err := os.Hostname()
	if err == nil {
		b.immutable["hostname"] = hostname
	}
}

func startReportTicker(rt core_runtime.Runtime, buffer *reportsBuffer) {
	go func() {
		err := buffer.dispatch(rt, pingHost, pingPort, "start")
		if err != nil {
			log.V(2).Info("Failed sending usage info", err)
		}
		for range time.Tick(time.Second * pingInterval) {
			err := buffer.dispatch(rt, pingHost, pingPort, "ping")
			if err != nil {
				log.V(2).Info("Failed sending usage info", err)
			}
		}
	}()
}

// Init core reports
func Init(rt core_runtime.Runtime, cfg kuma_cp.Config) {
	var buffer reportsBuffer
	buffer.immutable = make(map[string]string)
	buffer.mutable = make(map[string]string)

	buffer.initImmutable(rt)

	if cfg.Reports.Enabled {
		startReportTicker(rt, &buffer)
	}
}
