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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	thanosv1beta1 "github.com/orangesys/thanos-operator/api/v1beta1"
	"github.com/orangesys/thanos-operator/util"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ReceiverReconciler reconciles a Receiver object
type ReceiverReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=receivers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=receivers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *ReceiverReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("receiver", req.NamespacedName)

	// Fetch the receiver instance
	receiver := &thanosv1beta1.Receiver{}
	if err := r.Get(ctx, req.NamespacedName, receiver); err != nil {
		if ignoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch thanos receiver")
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
		util.SetService(service, receiver)
		return controllerutil.SetControllerReference(receiver, service, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// Generate StatefulSet
	ss := &appsv1.StatefulSet{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	_, err = ctrl.CreateOrUpdate(ctx, r.Client, ss, func() error {
		util.SetStatefulSet(
			ss,
			service,
			*receiver,
		)
		return controllerutil.SetControllerReference(receiver, ss, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	// Update Status
	ssNN := req.NamespacedName
	ssNN.Name = ss.Name
	if err := r.Get(ctx, ssNN, ss); err != nil {
		log.Error(err, "unable to fetch StatusfulSet", "namespaceName", ssNN)
		return ctrl.Result{}, err
	}
	receiver.Status.StatefulSetStatus = ss.Status

	serviceNN := req.NamespacedName
	serviceNN.Name = service.Name
	if err := r.Get(ctx, serviceNN, service); err != nil {
		log.Error(err, "unable to fetch Service", "namespaceName", serviceNN)
		return ctrl.Result{}, err
	}
	receiver.Status.ServiceStatus = service.Status

	err = r.Status().Update(ctx, receiver)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ReceiverReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&thanosv1beta1.Receiver{}).
		Owns(&appsv1.StatefulSet{}). // Generates StatefulSets
		Owns(&corev1.Service{}).     // Generates Services
		Complete(r)
}

func ignoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}
