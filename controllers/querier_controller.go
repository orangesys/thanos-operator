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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"

	thanosv1beta1 "github.com/orangesys/thanos-operator/api/v1beta1"
)

// QuerierReconciler reconciles a Querier object
type QuerierReconciler struct {
	client.Client
	Log      logr.Logger
	Recorder record.EventRecorder
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=queriers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=thanos.orangesys.io,resources=queriers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;create;update;patch;delete

func (r *QuerierReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("querier", req.NamespacedName)

	// Fetch the receiver instance
	querier := &thanosv1beta1.Querier{}
	if err := r.Get(ctx, req.NamespacedName, querier); err != nil {
		if ignoreNotFound(err) == nil {
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch thanos querier")
		return ctrl.Result{}, err
	}

	// Generate Service
	service := &corev1.Service{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, service, func() error {
		if ownerRef := ownedByOther(service, thanosv1beta1.GroupVersion, "Querier", req.Name); ownerRef != nil {
			log.Info("cowardly refusing to take ownership of somebody else's service", "owner", ownerRef)
			setCondition(&querier.Status.Conditions, thanosv1beta1.StatusCondition{
				Type:    "ServiceUpToDate",
				Status:  thanosv1beta1.ConditionStatusUnhealthy,
				Reason:  "OrphanService",
				Message: "Refusing to take ownership of somebody else's service",
			})
			return nil
		}
		makeService(service, service.Name)
		if err := ctrl.SetControllerReference(querier, service, r.Scheme); err != nil {
			return err
		}
		setCondition(&querier.Status.Conditions, thanosv1beta1.StatusCondition{
			Type:    "ServiceUpToDate",
			Status:  thanosv1beta1.ConditionStatusHealthy,
			Reason:  "EnsuredService",
			Message: "Ensured service was up to date",
		})
		return nil
	}); err != nil {
		log.Error(err, "unable to ensure service is correct")
		setCondition(&querier.Status.Conditions, thanosv1beta1.StatusCondition{
			Type:    "ServiceUpToDate",
			Status:  thanosv1beta1.ConditionStatusUnhealthy,
			Reason:  "UpdateError",
			Message: "Unable to fetch or update service",
		})
		if err := r.Status().Update(ctx, querier); err != nil {
			log.Error(err, "unable to update querier status")
		}
		return ctrl.Result{}, err
	}

	// Generate Deployment
	dm := &appsv1.Deployment{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, dm, func() error {
		setQuerierDeployment(
			dm,
			*querier,
		)

		if err := ctrl.SetControllerReference(querier, dm, r.Scheme); err != nil {
			return err
		}

		setCondition(&querier.Status.Conditions, thanosv1beta1.StatusCondition{
			Type:    "DeploymentUpToDate",
			Status:  thanosv1beta1.ConditionStatusHealthy,
			Reason:  "EnsuredDeployment",
			Message: "Ensured deployment was up to date",
		})
		return nil
	}); err != nil {
		log.Error(err, "unable to ensure deployment is correct")
		setCondition(&querier.Status.Conditions, thanosv1beta1.StatusCondition{
			Type:    "DeploymentUpToDate",
			Status:  thanosv1beta1.ConditionStatusUnhealthy,
			Reason:  "UpdateError",
			Message: "Unable to fetch or update deployment",
		})
		if err := r.Status().Update(ctx, querier); err != nil {
			log.Error(err, "unable to update querier status")
		}
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *QuerierReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&thanosv1beta1.Querier{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
