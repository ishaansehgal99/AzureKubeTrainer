package k8resources

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"

	trainingv1 "github.com/ishaansehgal99/AzureKubeTrainer/api/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GenerateServiceManifest(ctx context.Context, trainingJob *trainingv1.TrainingJob) *corev1.Service {
	klog.InfoS("GenerateServiceManifest", "workspace", klog.KObj(trainingJob), "serviceType", corev1.ServiceTypeLoadBalancer)

	selector := make(map[string]string)
	for k, v := range trainingJob.Spec.LabelSelector.MatchLabels {
		selector[k] = v
	}
	podNameForIndex0 := fmt.Sprintf("%s-0", trainingJob.Name)
	selector["statefulset.kubernetes.io/pod-name"] = podNameForIndex0

	return &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      trainingJob.Name,
			Namespace: trainingJob.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: trainingv1.GroupVersion.String(),
					Kind:       "TrainingJob",
					UID:        trainingJob.UID,
					Name:       trainingJob.Name,
				},
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(5000),
				},
			},
			Selector: selector,
		},
	}
}

func GenerateStatefulSetManifest(ctx context.Context, trainingJob *trainingv1.TrainingJob, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort, resourceRequirements corev1.ResourceRequirements,
	volumeMount []corev1.VolumeMount, tolerations []corev1.Toleration, volumes []corev1.Volume) *appsv1.StatefulSet {

	klog.InfoS("GenerateStatefulSetManifest", "workspace", klog.KObj(trainingJob), "image", imageName)

	// Gather label requirements from trianingJobs's label selector
	var labelRequirements []v1.LabelSelectorRequirement
	for key, value := range trainingJob.Spec.LabelSelector.MatchLabels {
		labelRequirements = append(labelRequirements, v1.LabelSelectorRequirement{
			Key:      key,
			Operator: v1.LabelSelectorOpIn,
			Values:   []string{value},
		})
	}

	return &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      trainingJob.Name,
			Namespace: trainingJob.Namespace,
			OwnerReferences: []v1.OwnerReference{
				{
					APIVersion: trainingv1.GroupVersion.String(),
					Kind:       "TrainingJob",
					UID:        trainingJob.UID,
					Name:       trainingJob.Name,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            lo.ToPtr(int32(replicas)),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector:            trainingJob.Spec.LabelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: trainingJob.Spec.LabelSelector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &v1.LabelSelector{
										MatchExpressions: labelRequirements,
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:         trainingJob.Name,
							Image:        imageName,
							Command:      commands,
							Resources:    resourceRequirements,
							Ports:        containerPorts,
							VolumeMounts: volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
		VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "azure-blob-storage",
					Annotations: map[string]string{
						"volume.beta.kubernetes.io/storage-class": "azureblob-nfs-premium",
					},
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: "blob-storage",
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteMany,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("5Gi"),
						},
					},
				},
			},
		},
	}
}
