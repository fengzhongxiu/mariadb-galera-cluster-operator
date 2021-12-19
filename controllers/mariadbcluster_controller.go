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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"mariadb-galera-cluster-operator.domain/utils"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mariadbv1 "mariadb-galera-cluster-operator.domain/api/v1"
)

// MariaDBClusterReconciler reconciles a MariaDBCluster object
type MariaDBClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mariadb.mariadb-galera-cluster-operator.domain,resources=mariadbclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mariadb.mariadb-galera-cluster-operator.domain,resources=mariadbclusters/status,verbs=get;update;patch

func (r *MariaDBClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("mariadbcluster", req.NamespacedName)

	// your logic here
	// 获取当前 Mariadb 集群实例
	var mariaDBCluster mariadbv1.MariaDBCluster
	if err := r.Client.Get(ctx, req.NamespacedName, &mariaDBCluster); err != nil {
		// EtcdCluster was deleted，Ignore
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 创建或更新 Service
	var svc corev1.Service
	svc.Name = mariaDBCluster.Name
	svc.Namespace = mariaDBCluster.Namespace
	or, err := ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
		// 调谐必须在这个函数中去实现
		utils.MutateHeadlessSVC(&mariaDBCluster, &svc)
		return controllerutil.SetControllerReference(&mariaDBCluster, &svc, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "Service", or)


	// CreateOrUpdate StatefulSet
	var sts appsv1.StatefulSet
	sts.Name = mariaDBCluster.Name
	sts.Namespace = mariaDBCluster.Namespace
	or, err = ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
		// 调谐必须在这个函数中去实现
		utils.MutateStatefulSet(&mariaDBCluster, &sts)
		return controllerutil.SetControllerReference(&mariaDBCluster, &sts, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	log.Info("CreateOrUpdate", "StatefulSet", or)

	return ctrl.Result{}, nil
}

func (r *MariaDBClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mariadbv1.MariaDBCluster{}).
		Complete(r)
}
