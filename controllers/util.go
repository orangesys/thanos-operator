package controllers

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	thanosv1beta1 "github.com/orangesys/thanos-operator/api/v1beta1"
)

const (
	governingServiceName = "thanos"
	defaultThanosVersion = "v0.5.0"
	defaultRetetion      = "24h"
	receiveStorage       = "2Gi"
	receiverDir          = "/thanos-receive"
	secretsDir           = "/etc/thanos/secrets/"
	sSetInputHashName    = "prometheus-operator-input-hash"
)

var (
	miniReplicas                int32 = 1
	gracePeriodTerm             int64 = 10
	managedByOperatorLabel            = "managed-by"
	managedByOperatorLabelValue       = "thanos-operator"
	managedByOperatorLabels           = map[string]string{
		managedByOperatorLabel: managedByOperatorLabelValue,
	}

	probeTimeoutSeconds int32 = 3
)

// setStoreDeployment set fields on appsv1.Depployment pointer generated
func setStoreDeployment(
	dm *appsv1.Deployment,
	service *corev1.Service,
	t thanosv1beta1.Store,
) {
	t = *t.DeepCopy()

	podLabels := map[string]string{
		"app":    "store",
		"thanos": t.Name,
	}
	if t.Spec.Resources.Requests == nil {
		t.Spec.Resources.Requests = corev1.ResourceList{}
	}

	_, memoryRequestFound := t.Spec.Resources.Requests[corev1.ResourceMemory]
	memoryLimit, memoryLimitFound := t.Spec.Resources.Limits[corev1.ResourceMemory]
	if !memoryRequestFound {
		defaultMemoryRequest := resource.MustParse("1Gi")
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		if memoryLimitFound && compareResult <= 0 {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = memoryLimit
		} else {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = defaultMemoryRequest
		}
	}

	podAnnotations := map[string]string{}

	if t.Spec.PodMetadata != nil {
		if t.Spec.PodMetadata.Labels != nil {
			for k, v := range t.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if t.Spec.PodMetadata.Annotations != nil {
			for k, v := range t.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	dm.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: podLabels,
	}
	dm.Spec.Replicas = &miniReplicas

	thanosArgs := []string{
		"store",
		fmt.Sprintf("--index-cache-size=%s", t.Spec.IndexCacheSize),
		fmt.Sprintf("--chunk-pool-size=%s", t.Spec.ChunkPoolSize),
		fmt.Sprintf("--data-dir=%s", t.Spec.DataDir),
		fmt.Sprintf("--objstore.config=type: %s\nconfig:\n  bucket: \"%s\"", t.Spec.ObjectStorageType, t.Spec.BucketName),
	}
	if t.Spec.LogLevel != "" && t.Spec.LogLevel != "info" {
		thanosArgs = append(thanosArgs, fmt.Sprintf("--log.level=%s", t.Spec.LogLevel))
	}

	env := []corev1.EnvVar{
		{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: secretsDir + t.Spec.SecretName + ".json",
		},
	}

	ports := []corev1.ContainerPort{
		{
			ContainerPort: 10902,
			Name:          "http",
		},
		{
			ContainerPort: 10901,
			Name:          "grpc",
		},
	}

	// mount to pod
	volumemounts := []corev1.VolumeMount{
		{
			Name:      "google-cloud-key",
			MountPath: secretsDir,
		},
	}

	containers := []corev1.Container{
		{
			Name:         "store",
			Image:        *t.Spec.Image,
			Args:         thanosArgs,
			Env:          env,
			Ports:        ports,
			VolumeMounts: volumemounts,
		},
	}
	volumes := []corev1.Volume{
		{
			Name: "google-cloud-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: t.Spec.SecretName,
				},
			},
		},
	}

	podspec := corev1.PodSpec{
		TerminationGracePeriodSeconds: &gracePeriodTerm,
		Containers:                    containers,
		Volumes:                       volumes,
	}

	dm.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: dm.Spec.Selector.MatchLabels,
		},
		Spec: podspec,
	}
}

// setQuerierDeployment set fields on a appsv1.Depployment pointer generated
func setQuerierDeployment(
	dm *appsv1.Deployment,
	service *corev1.Service,
	t thanosv1beta1.Querier,
) {
	t = *t.DeepCopy()

	podLabels := map[string]string{
		"app":    "querier",
		"thanos": t.Name,
	}
	if t.Spec.Resources.Requests == nil {
		t.Spec.Resources.Requests = corev1.ResourceList{}
	}

	_, memoryRequestFound := t.Spec.Resources.Requests[corev1.ResourceMemory]
	memoryLimit, memoryLimitFound := t.Spec.Resources.Limits[corev1.ResourceMemory]
	if !memoryRequestFound {
		defaultMemoryRequest := resource.MustParse("1Gi")
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		if memoryLimitFound && compareResult <= 0 {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = memoryLimit
		} else {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = defaultMemoryRequest
		}
	}

	podAnnotations := map[string]string{}

	if t.Spec.PodMetadata != nil {
		if t.Spec.PodMetadata.Labels != nil {
			for k, v := range t.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if t.Spec.PodMetadata.Annotations != nil {
			for k, v := range t.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	dm.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: podLabels,
	}
	dm.Spec.Replicas = &miniReplicas

	thanosArgs := []string{
		"query",
		fmt.Sprintf("--query.replica-label=%s", t.Spec.ReplicaLabel),
		// fmt.Sprintf("--store=dnssrv+%s", t.Spec.StoreDNS),
	}
	if t.Spec.LogLevel != "" && t.Spec.LogLevel != "info" {
		thanosArgs = append(thanosArgs, fmt.Sprintf("--log.level=%s", t.Spec.LogLevel))
	}

	ports := []corev1.ContainerPort{
		{
			ContainerPort: 10902,
			Name:          "http",
		},
		{
			ContainerPort: 10901,
			Name:          "grpc",
		},
	}

	containers := []corev1.Container{
		{
			Name:  "querier",
			Image: *t.Spec.Image,
			Args:  thanosArgs,
			Ports: ports,
		},
	}
	podspec := corev1.PodSpec{
		TerminationGracePeriodSeconds: &gracePeriodTerm,
		Containers:                    containers,
	}

	dm.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: dm.Spec.Selector.MatchLabels,
		},
		Spec: podspec,
	}
}

// SetStatefulSetService set filds on a appsv1.StatefulSet pointer generated and
// the Service object for the Thanos instance
// SetStatefulSetFields sets fields on a appsv1.StatefulSet pointer generated for the Thanos instance
// object: Thanos instance
// replicas: the number of replicas for the Thanos instance
// storage: the size of the storage for the Thanos instance (e.g. 2Gi)
func setReceiverStatefulSet(
	ss *appsv1.StatefulSet,
	service *corev1.Service,
	t thanosv1beta1.Receiver,
) {
	t = *t.DeepCopy()

	podLabels := map[string]string{}
	rl := corev1.ResourceList{}

	switch {
	case strings.HasPrefix(t.Name, "receiver"):
		if t.Spec.Storage == "" {
			t.Spec.Storage = receiveStorage
		}
		podLabels["app"] = "receiver"
		podLabels["thanos"] = t.Name
		podLabels["thanos-store-api"] = "true"

		rl["storage"] = resource.MustParse(t.Spec.Storage)

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

	if t.Spec.Resources.Requests == nil {
		t.Spec.Resources.Requests = corev1.ResourceList{}
	}

	_, memoryRequestFound := t.Spec.Resources.Requests[corev1.ResourceMemory]
	memoryLimit, memoryLimitFound := t.Spec.Resources.Limits[corev1.ResourceMemory]
	if !memoryRequestFound {
		defaultMemoryRequest := resource.MustParse("1Gi")
		compareResult := memoryLimit.Cmp(defaultMemoryRequest)
		if memoryLimitFound && compareResult <= 0 {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = memoryLimit
		} else {
			t.Spec.Resources.Requests[corev1.ResourceMemory] = defaultMemoryRequest
		}
	}

	podAnnotations := map[string]string{}

	if t.Spec.PodMetadata != nil {
		if t.Spec.PodMetadata.Labels != nil {
			for k, v := range t.Spec.PodMetadata.Labels {
				podLabels[k] = v
			}
		}
		if t.Spec.PodMetadata.Annotations != nil {
			for k, v := range t.Spec.PodMetadata.Annotations {
				podAnnotations[k] = v
			}
		}
	}

	ss.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: podLabels,
	}
	ss.Spec.ServiceName = service.Name
	ss.Spec.Replicas = &miniReplicas

	podspec, err := makePodSpec(t)
	if err != nil {
		return
	}

	ss.Spec.Template = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: ss.Spec.Selector.MatchLabels,
		},
		Spec: *podspec,
	}

}

// makePodSpec  is create spec
func makePodSpec(t thanosv1beta1.Receiver) (*corev1.PodSpec, error) {

	if t.Spec.ReceivePrefix == "" {
		t.Spec.ReceivePrefix = receiverDir
	}
	if t.Spec.Retention == "" {
		t.Spec.Retention = defaultRetetion
	}
	// TODO set args to spec
	thanosArgs := []string{
		"receive",
		fmt.Sprintf("--tsdb.path=%s", t.Spec.ReceivePrefix),
		fmt.Sprintf("--tsdb.retention=%s", t.Spec.Retention),
		fmt.Sprintf("--labels=receive=\"%s\"", t.Spec.ReceiveLables),
		fmt.Sprintf("--objstore.config=type: %s\nconfig:\n  bucket: \"%s\"", t.Spec.ObjectStorageType, t.Spec.BucketName),
	}
	if t.Spec.LogLevel != "" && t.Spec.LogLevel != "info" {
		thanosArgs = append(thanosArgs, fmt.Sprintf("--log.level=%s", t.Spec.LogLevel))
	}
	env := []corev1.EnvVar{
		{
			Name:  "GOOGLE_APPLICATION_CREDENTIALS",
			Value: secretsDir + t.Spec.SecretName + ".json",
		},
	}

	ports := []corev1.ContainerPort{
		{
			ContainerPort: 10902,
			Name:          "http",
		},
		{
			ContainerPort: 10901,
			Name:          "grpc",
		},
	}

	if strings.Contains(t.Name, "receiver") {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: 19291,
			Name:          "receive",
		})
	}

	// mount to pod
	volumemounts := []corev1.VolumeMount{
		{
			Name:      "thanos-persistent-storage",
			MountPath: t.Spec.Retention,
		},
		{
			Name:      "google-cloud-key",
			MountPath: secretsDir,
		},
	}

	containers := []corev1.Container{
		{
			Name:         "receiver",
			Image:        *t.Spec.Image,
			Args:         thanosArgs,
			Env:          env,
			Ports:        ports,
			VolumeMounts: volumemounts,
		},
	}

	// Need create json from gcp iam
	// https://github.com/orangesys/blueprint/tree/master/prometheus-thanos
	// kubectl create secret generic ${SERVICE_ACCOUNT_NAME} --from-file=${SERVICE_ACCOUNT_NAME}.json=${SERVICE_ACCOUNT_NAME}.json
	// secret name is thanos-demo-gcs
	// TODO setting secret name with spec
	volumes := []corev1.Volume{
		{
			Name: "google-cloud-key",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: t.Spec.SecretName,
				},
			},
		},
	}

	return &corev1.PodSpec{
		TerminationGracePeriodSeconds: &gracePeriodTerm,
		Containers:                    containers,
		Volumes:                       volumes,
	}, nil
}

func makeService(service *corev1.Service, role string) {
	if strings.Contains(role, "receiver") {
		service.Labels = map[string]string{
			"service": "receiver",
			"thanos":  role,
		}
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port: 19291,
				Name: "receive",
			},
			{
				Port: 10902,
				Name: "http",
			},
			{
				Port: 10901,
				Name: "grpc",
			},
		}
	}
	if strings.Contains(role, "querier") {
		service.Labels = map[string]string{
			"service": "querier",
			"thanos":  role,
		}
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port: 10902,
				Name: "http",
			},
			{
				Port: 10901,
				Name: "grpc",
			},
		}
	}
	if strings.Contains(role, "store") {
		service.Labels = map[string]string{
			"service": "store",
			"thanos":  role,
		}
		service.Spec.Ports = []corev1.ServicePort{
			{
				Port: 10902,
				Name: "http",
			},
			{
				Port: 10901,
				Name: "grpc",
			},
		}
	}
	service.Spec.Selector = map[string]string{"thanos": role}
}

func ignoreNotFound(err error) error {
	if errors.IsNotFound(err) {
		return nil
	}
	return err
}
