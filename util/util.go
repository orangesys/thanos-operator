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

package util

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SetStatefulSetFields sets fields on a appsv1.StatefulSet pointer generated for the Thanos instance
// object: Thanos instance
// replicas: the number of replicas for the Thanos instance
// storage: the size of the storage for the Thanos instance (e.g. 2Gi)
func SetStatefulSetFields(ss *appsv1.StatefulSet, service *corev1.Service, thanos metav1.Object, storage *string) {
	gracePeriodTerm := int64(10)
	replicas := int32(1)

	if storage == nil {
		s := "2Gi"
		storage = &s
	}

	copyLabels := thanos.GetLabels()
	if copyLabels == nil {
		copyLabels = map[string]string{}
	}

	labels := map[string]string{}
	for k, v := range copyLabels {
		labels[k] = v
	}
	labels["receiver-statefuleset"] = thanos.GetName()

	rl := corev1.ResourceList{}
	rl["storage"] = resource.MustParse(*storage)

	ss.Labels = labels
	ss.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{"receiver-statefulset": thanos.GetName()},
	}
	ss.Spec.ServiceName = service.Name
	ss.Spec.Replicas = &replicas
	ss.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: ss.Spec.Selector.MatchLabels,
		},

		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: &gracePeriodTerm,
			Containers: []corev1.Container{
				{
					Name:  "thanos",
					Image: "improbable/thanos:v0.5.0",
					Args:  []string{"receive", "--log.level=debug", "--tsdb.path=/thanos-receive", "--tsdb.retention=3h", "--labels=receive=\"true\""},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 19291,
							Name:          "receive",
						},
						{
							ContainerPort: 10901,
							Name:          "grpc",
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: "thanos-persistent-storage", MountPath: "/thanos-receive"}},
				},
			},
		},
	}
	ss.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "thanos-persistent-storage"},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				Resources: corev1.ResourceRequirements{
					Requests: rl,
				},
			},
		},
	}
}

// DeploymentFields set field on the thanos object
// func DeploymentFields(dm *appsv1.Deployment, service *corev1.Service, thanos metav1.Object, cont corev1.Container, storage *string) {
// 	gracePeriodTerm := int64(10)
// 	replicas := int32(1)

// 	if storage == nil {
// 		s := "2Gi"
// 		storage = &s
// 	}

// 	copyLabels := thanos.GetLabels()
// 	if copyLabels == nil {
// 		copyLabels = map[string]string{}
// 	}

// 	labels := map[string]string{}
// 	for k, v := range copyLabels {
// 		labels[k] = v
// 	}

// 	labels["receiver-deployment"] = thanos.GetName()

// 	rl := corev1.ResourceList{}
// 	rl["storage"] = resource.MustParse(*storage)

// 	dm.Labels = labels
// 	dm.Spec.Selector = &metav1.LabelSelector{
// 		MatchLabels: map[string]string{"receiver-deployment": thanos.GetName()},
// 	}

// 	dm.Spec.Replicas = &replicas
// 	dm.Spec.Template = corev1.PodTemplateSpec{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Labels: dm.Spec.Selector.MatchLabels,
// 		},

// 		Spec: corev1.PodSpec{
// 			TerminationGracePeriodSeconds: &gracePeriodTerm,
// 			Containers: []corev1.Container{
// 				{
// 					Name:         "thanos",
// 					Image:        "improbable/thanos:v0.5.0-rc.0",
// 					Args:         []string{"receive", "--log.level=debug", "--tsdb.path=/thanos-receive", "--tsdb.retention=3h", "--labels=receive=\"true\""},
// 					Ports:        []corev1.ContainerPort{{ContainerPort: 19291}},
// 					VolumeMounts: []corev1.VolumeMount{{Name: "receiver-persistent-storage", MountPath: "/thanos-receive"}},
// 				},
// 			},
// 		},
// 	}
// }

// SetServiceFields sets fields on the Service object
func SetServiceFields(service *corev1.Service, thanos metav1.Object) {
	copyLabels := thanos.GetLabels()
	if copyLabels == nil {
		copyLabels = map[string]string{}
	}
	labels := map[string]string{}
	for k, v := range copyLabels {
		labels[k] = v
	}
	service.Labels = labels

	service.Spec.Ports = []corev1.ServicePort{
		{
			Port: 19291,
			Name: "receive",
		},
		{
			Port: 10901,
			Name: "grpc",
		},
		// {Port: 19291, TargetPort: intstr.IntOrString{IntVal: 19291, Type: intstr.Int}},
		// {Port: 10901, TargetPort: intstr.IntOrString{IntVal: 10901, Type: intstr.Int}},
	}
	service.Spec.Selector = map[string]string{"receiver-statefulset": thanos.GetName()}
}
