/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/plugins/bootstrap/k8s/cache/internal"
	"github.com/pkg/errors"
)

var (
	log               = core.Log.WithName("object-cache")
	defaultResyncTime = 10 * time.Hour
)

// New initializes and returns a new Cache.
func New(config *rest.Config, opts cache.Options) (cache.Cache, error) {
	opts, err := defaultOpts(config, opts)
	if err != nil {
		return nil, err
	}
	im := internal.NewInformersMap(config, opts.Scheme, opts.Mapper, *opts.Resync, opts.Namespace)
	return &informerCache{InformersMap: im}, nil
}

func defaultOpts(config *rest.Config, opts cache.Options) (cache.Options, error) {
	// Use the default Kubernetes Scheme if unset
	if opts.Scheme == nil {
		opts.Scheme = scheme.Scheme
	}

	// Construct a new Mapper if unset
	if opts.Mapper == nil {
		httpClient, err := rest.HTTPClientFor(config)
		if err != nil {
			return cache.Options{}, errors.Wrap(err, "failed to create HTTP client from Manager config")
		}
		opts.Mapper, err = apiutil.NewDiscoveryRESTMapper(config, httpClient)
		if err != nil {
			log.WithName("setup").Error(err, "Failed to get API Group-Resources")
			return opts, fmt.Errorf("could not create RESTMapper from config")
		}
	}

	// Default the resync period to 10 hours if unset
	if opts.Resync == nil {
		opts.Resync = &defaultResyncTime
	}
	return opts, nil
}
