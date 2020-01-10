package xds

import (
	"context"
	"io"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
	mads_client "github.com/Kong/kuma/pkg/mads/client"
)

type DiscoveryConfig struct {
	ServerURL  string
	ClientName string
}

type discoverer struct {
	log           logr.Logger
	config        DiscoveryConfig
	oldSourceList map[string]bool
}

func NewDiscoverer(config DiscoveryConfig, log logr.Logger) (*discoverer, error) {
	cd := &discoverer{
		log:           log,
		config:        config,
		oldSourceList: make(map[string]bool),
	}
	return cd, nil
}

// Run implements discovery.Discoverer interface.
func (d *discoverer) Run(ctx context.Context, ch chan<- []*targetgroup.Group) {
	// notice that Prometheus discovery.Discoverer abstraction doesn't allow failures,
	// so we must ensure that xDS client is up-and-running all the time.
	for streamID := uint64(1); ; streamID++ {
		errCh := make(chan error, 1)
		go func(errCh chan<- error) {
			defer close(errCh)
			// recover from a panic
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok {
						errCh <- err
					} else {
						errCh <- errors.Errorf("%v", e)
					}
				}
			}()
			errCh <- d.stream(ctx, streamID, ch)
		}(errCh)
		select {
		case <-ctx.Done():
			d.log.Info("done")
			break
		case err := <-errCh:
			if err != nil {
				d.log.WithValues("streamID", streamID).Error(err, "xDS stream terminated with an error")
			}
		}
	}
}

func (d *discoverer) stream(ctx context.Context, streamID uint64, ch chan<- []*targetgroup.Group) (errs error) {
	log := d.log.WithValues("streamID", streamID)

	log.Info("creating a gRPC client for Monitoring Assignment Discovery Service (MADS) server ...")
	client, err := mads_client.New(d.config.ServerURL)
	if err != nil {
		return errors.Wrap(err, "failed to connect to gRPC server")
	}
	defer func() {
		log.Info("closing a connection ...")
		if err := client.Close(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "failed to close a connection"))
		}
	}()

	log.Info("starting an xDS stream ...")
	stream, err := client.StartStream()
	if err != nil {
		return errors.Wrap(err, "failed to start an xDS stream")
	}
	defer func() {
		log.Info("closing an xDS stream ...")
		if err := stream.Close(); err != nil {
			errs = multierr.Append(errs, errors.Wrapf(err, "failed to close an xDS stream"))
		}
	}()

	log.Info("sending first discovery request on a new xDS stream ...")
	err = stream.RequestAssignments(d.config.ClientName)
	if err != nil {
		return errors.Wrap(err, "failed to send a discovery request")
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		log.Info("waiting for a discovery response ...")
		assignments, err := stream.WaitForAssignments()
		if err != nil {
			return errors.Wrap(err, "failed to receive a discovery response")
		}
		log.Info("received monitoring assignments", "len", len(assignments))
		log.V(1).Info("received monitoring assignments", "assignments", assignments)

		tgs, newSourceList := d.convertAll(assignments)
		ch <- tgs
		d.oldSourceList = newSourceList

		if err := stream.ACK(); err != nil {
			if err == io.EOF {
				break
			}
			return errors.Wrap(err, "failed to ACK a discovery response")
		}
	}

	return nil
}

func (d *discoverer) convertAll(assignments []*observability_proto.MonitoringAssignment) ([]*targetgroup.Group, map[string]bool) {
	var tgs []*targetgroup.Group
	newSourceList := make(map[string]bool)
	for _, assignment := range assignments {
		tg := d.convertOne(assignment)
		tgs = append(tgs, tg)
		newSourceList[tg.Source] = true
	}

	// when targetGroup disappear, we should send an update with an empty targetList
	for key := range d.oldSourceList {
		if !newSourceList[key] {
			tgs = append(tgs, &targetgroup.Group{
				Source: key,
			})
		}
	}
	return tgs, newSourceList
}

func (d *discoverer) convertOne(assignment *observability_proto.MonitoringAssignment) *targetgroup.Group {
	tg := &targetgroup.Group{
		Source: assignment.Name,
	}
	tg.Labels = d.convertLabels(assignment.Labels)
	for _, target := range assignment.Targets {
		tg.Targets = append(tg.Targets, d.convertLabels(target.Labels))
	}
	return tg
}

func (d *discoverer) convertLabels(labels map[string]string) model.LabelSet {
	labelSet := model.LabelSet{}
	for key, value := range labels {
		name := strutil.SanitizeLabelName(key)
		labelSet[model.LabelName(name)] = model.LabelValue(value)
	}
	return labelSet
}
