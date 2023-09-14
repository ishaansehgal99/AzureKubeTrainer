/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"
	"github.com/ishaansehgal99/AzureKubeTrainer/pkg/k8resources"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strconv"
	"time"

	trainingv1 "github.com/ishaansehgal99/AzureKubeTrainer/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var torchRunParams = map[string]string{}

// TrainingJobReconciler reconciles a TrainingJob object
type TrainingJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func getResource(ctx context.Context, name, namespace string, kubeClient client.Client, resource client.Object) error {
	// Log the retrieval attempt.
	resourceType := fmt.Sprintf("%T", resource)
	klog.InfoS(fmt.Sprintf("Get%s", resourceType), "resourceName", name, "resourceNamespace", namespace)

	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, resource, &client.GetOptions{})
	})

	return err
}

func createResource(ctx context.Context, resource client.Object, kubeClient client.Client) error {
	// Log the creation attempt.
	klog.InfoS("CreateResource", "resource", klog.KObj(resource))
	// Create the resource.
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, resource, &client.CreateOptions{})
	})
}

//func getService(ctx context.Context, name, namespace string, kubeClient client.Client) (*v1.Service, error) {
//	klog.InfoS("GetService", "serviceName", name, "serviceNamespace", namespace)
//
//	svc := &v1.Service{}
//	err := retry.OnError(retry.DefaultBackoff, func(err error) bool {
//		return true
//	}, func() error {
//		return kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, svc, &client.GetOptions{})
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	return svc, nil
//}

func createService(ctx context.Context, serviceObj *v1.Service, kubeClient client.Client) error {
	klog.InfoS("CreateService", "service", klog.KObj(serviceObj))
	return retry.OnError(retry.DefaultBackoff, func(err error) bool {
		return true
	}, func() error {
		return kubeClient.Create(ctx, serviceObj, &client.CreateOptions{})
	})
}

func buildCommand(baseCommand string, torchRunParams map[string]string) []string {
	var updatedBaseCommand string
	for key, value := range torchRunParams {
		updatedBaseCommand = fmt.Sprintf("%s --%s=%s", baseCommand, key, value)
	}

	commands := []string{
		"/bin/sh",
		"-c",
		updatedBaseCommand,
	}

	return commands
}

func (r *TrainingJobReconciler) createStatefulsetIfNil(ctx context.Context, trainingObj *trainingv1.TrainingJob) error {
	// Check if statefulset already exists
	err := getResource(ctx, trainingObj.Name, trainingObj.Namespace, r.Client, trainingObj)
	if err != nil && !errors.IsNotFound(err) {
		klog.ErrorS(err, "failed to check for existing TrainingJob", "TrainingJob", trainingObj)
		return err
	}

	if trainingObj.GetName() != "" && trainingObj.GetNamespace() != "" {
		klog.InfoS("a trainingObj already exists", "TrainingJob", klog.KObj(trainingObj))
		return nil
	}

	// Get existing service
	existingSVC := &v1.Service{}
	err = getResource(ctx, trainingObj.Name, trainingObj.Namespace, r.Client, existingSVC)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	replicas := trainingObj.Spec.Replicas
	procPerNode := trainingObj.Spec.ProcPerNode
	torchRunParams["nnodes"] = strconv.Itoa(replicas)
	torchRunParams["node_rank"] = "$(echo $HOSTNAME | grep -o '[^-]*$')"
	torchRunParams["nproc_per_node"] = strconv.Itoa(procPerNode)
	torchRunParams["master_addr"] = existingSVC.Spec.LoadBalancerIP
	torchRunParams["master_port"] = "80"

	commands := buildCommand("torchrun training.py", torchRunParams)

	resourceRequirements := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("4"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceName("nvidia.com/gpu"): resource.MustParse("4"),
			corev1.ResourceEphemeralStorage:       resource.MustParse("300Gi"),
		},
	}

	volume := []corev1.Volume{
		{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium: "Memory",
				},
			},
		},
	}

	volumeMount := []corev1.VolumeMount{}
	if len(volume) != 0 {
		volumeMount = append(volumeMount, corev1.VolumeMount{
			Name:      volume[0].Name,
			MountPath: "/dev/shm",
		})
	}

	containerPorts := []corev1.ContainerPort{
		{
			ContainerPort: int32(5000),
		},
	}

	tolerations := []corev1.Toleration{
		{
			Effect:   corev1.TaintEffectNoSchedule,
			Operator: corev1.TolerationOpEqual,
			Key:      "gpu",
		},
		{
			Effect: corev1.TaintEffectNoSchedule,
			Value:  "gpu",
			Key:    "sku",
		},
	}

	depObj := k8resources.GenerateStatefulSetManifest(ctx, trainingObj, PresetSetModelllama2CChatImage, /*todo put image here*/
		replicas, commands, containerPorts, resourceRequirements, volumeMount, tolerations, volume)

	if err := createResource(ctx, depObj, r.Client); err != nil {
		return err
	}

	return nil
}

func (r *TrainingJobReconciler) createServiceIfNil(ctx context.Context, trainingObj *trainingv1.TrainingJob) error {
	// Check if service already exists
	existingSVC := &v1.Service{}
	err := getResource(ctx, trainingObj.Name, trainingObj.Namespace, r.Client, existingSVC)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if existingSVC != nil {
		klog.InfoS("a service already exists for TrainingJob", "TrainingJob", klog.KObj(trainingObj), "serviceType", corev1.ServiceTypeLoadBalancer)
		return nil
	}

	// If service doesn't exist, create it
	serviceObj := k8resources.GenerateServiceManifest(ctx, trainingObj)
	err = createService(ctx, serviceObj, r.Client)
	klog.InfoS("a service has been created for TrainingJob", "TrainingJob", klog.KObj(trainingObj))
	return err
}

//+kubebuilder:rbac:groups=training.azure.kube.trainer,resources=trainingjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=training.azure.kube.trainer,resources=trainingjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=training.azure.kube.trainer,resources=trainingjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TrainingJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *TrainingJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	trainingObj := &trainingv1.TrainingJob{}
	if err := r.Client.Get(ctx, req.NamespacedName, trainingObj); err != nil {
		if !errors.IsNotFound(err) {
			klog.ErrorS(err, "failed to get trainingObj", "TrainingJob", req.Name)
		}
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	klog.InfoS("Reconciling", "TrainingJob", req.NamespacedName)

	return r.createTrainingJob(ctx, trainingObj)
}

func (r *TrainingJobReconciler) createTrainingJob(ctx context.Context, trainingObj *trainingv1.TrainingJob) (ctrl.Result, error) {
	klog.InfoS("createTrainingJob", "TrainingJob", klog.KObj(trainingObj))

	err := r.createServiceIfNil(ctx, trainingObj)
	if err != nil {
		klog.InfoS("failed to ensure TrainingJob Service was made", "TrainingJob", klog.KObj(trainingObj))
		return reconcile.Result{RequeueAfter: time.Duration(10) * time.Second}, err
	}

	err = r.createStatefulsetIfNil(ctx, trainingObj)
	if err != nil {
		klog.InfoS("failed to ensure TrainingJob Statefulset was made", "TrainingJob", klog.KObj(trainingObj))
		return reconcile.Result{RequeueAfter: time.Duration(10) * time.Second}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TrainingJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1.TrainingJob{}).
		Complete(r)
}
