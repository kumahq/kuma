/*
Copyright 2019 The Kubernetes Authors.

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

// Package cache is a copy of the package sigs.k8s.io/controller-runtime/pkg/cache
// but without deepCopy on Get and List methods. We actively call Get/List from multiple
// goroutines and everywhere treat results as immutable. So apparently deepCopy doesn't make
// any sense to us and leads to extra memory usage.

// Package cache provides object caches that act as caching client.Reader
// instances and help drive Kubernetes-object-based event handlers.
package cache
