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
)

// DataplaneSpec defines the desired state of Dataplane
type DataplaneSpec = map[string]interface{}

// Dataplane is the Schema for the Dataplane API
type Dataplane struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Mesh string        `json:"mesh,omitempty"`
	Spec DataplaneSpec `json:"spec,omitempty"`
}

// DataplaneList contains a list of Dataplane
type DataplaneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataplane `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dataplane{}, &DataplaneList{})
}
