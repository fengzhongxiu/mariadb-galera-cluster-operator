/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MariaDBClusterSpec defines the desired state of MariaDBCluster
type MariaDBClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// mariadb的副本数
	Size *int32 `json:"size,omitempty"`
}

// MariaDBClusterStatus defines the observed state of MariaDBCluster
type MariaDBClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// mariadb的副本数
	Size *int32 `json:"size,omitempty"`
}

// +kubebuilder:object:root=true

// MariaDBCluster is the Schema for the mariadbclusters API
type MariaDBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MariaDBClusterSpec   `json:"spec,omitempty"`
	Status MariaDBClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MariaDBClusterList contains a list of MariaDBCluster
type MariaDBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MariaDBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MariaDBCluster{}, &MariaDBClusterList{})
}
