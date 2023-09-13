package k8resources

import (
	"context"

	trainingv1 "github.com/ishaansehgal99/AzureKubeTrainer/nccl/api/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func GenerateStatefulSetManifest(ctx context.Context, trainingJobObj *trainingv1.trainingJob, imageName string,
	replicas int, commands []string, containerPorts []corev1.ContainerPort,
	livenessProbe, readinessProbe *corev1.Probe, resourceRequirements corev1.ResourceRequirements,
	volumeMount []corev1.VolumeMount, tolerations []corev1.Toleration, volumes []corev1.Volume) *appsv1.StatefulSet {

	klog.InfoS("GenerateStatefulSetManifest", "workspace", klog.KObj(trainingJob), "image", imageName)

	// Gather label requirements from trianingJobs's label selector
	var labelRequirements []v1.LabelSelectorRequirement
	for key, value := range trainingJob.Resource.LabelSelector.MatchLabels {
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
					Kind:       "Workspace",
					UID:        trainingJob.UID,
					Name:       trainingJob.Name,
				},
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            lo.ToPtr(int32(replicas)),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector:            trainingJob.Resource.LabelSelector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: trainingJob.Resource.LabelSelector.MatchLabels,
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
							Name:           trainingJob.Name,
							Image:          imageName,
							Command:        commands,
							Resources:      resourceRequirements,
							LivenessProbe:  livenessProbe,
							ReadinessProbe: readinessProbe,
							Ports:          containerPorts,
							VolumeMounts:   volumeMount,
						},
					},
					Tolerations: tolerations,
					Volumes:     volumes,
				},
			},
		},
	}
}
