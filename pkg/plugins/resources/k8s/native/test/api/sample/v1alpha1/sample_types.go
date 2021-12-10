/*
Copyright 2019 Kuma authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	test_proto "github.com/kumahq/kuma/pkg/test/apis/sample/v1alpha1"
)

// TrafficRouteSpec defines the desired state of SampleTrafficRoute
type TrafficRouteSpec = *test_proto.TrafficRoute

// SampleTrafficRoute is the Schema for the proxytemplates API
type SampleTrafficRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Mesh              string `json:"mesh"`

	Spec TrafficRouteSpec `json:"spec,omitempty"`
}

// SampleTrafficRouteList contains a list of SampleTrafficRoute
type SampleTrafficRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SampleTrafficRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SampleTrafficRoute{}, &SampleTrafficRouteList{})
}
