/*
Copyright 2019 Gavin Zhou.

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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	thanosv1beta1 "github.com/orangesys/thanos-operator/api/v1beta1"
)

// StoreReconciler reconciles a Store object
type StoreReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=stores,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=stores/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete

func (r *StoreReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("store", req.NamespacedName)

	store := &thanosv1beta1.Store{}
	if err := r.Get(ctx, req.NamespacedName, store); err != nil {
		if ignoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch thanos store")
		return ctrl.Result{}, err
	}

	// Generate Service
	service := &corev1.Service{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}
	_, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		makeService(service, service.Name)
		return controllerutil.SetControllerReference(store, service, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}
	// Generate Deployment
	dm := &appsv1.Deployment{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, dm, func() error {
		setStoreDeployment(
			dm,
			*store,
		)
		return controllerutil.SetControllerReference(store, dm, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	// Update Status

	return ctrl.Result{}, nil
}

func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&thanosv1beta1.Store{}).
		Complete(r)
}
