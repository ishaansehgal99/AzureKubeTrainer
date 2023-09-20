# Setup
##  Deploy
```shell
git clone https://github.com/kubeflow/training-operator
cd training-operator
Make deploy
```
## Play with the example
* Intall GPUs drivers on Azure Kubernetes Service (AKS) 
  * https://learn.microsoft.com/en-us/azure/aks/gpu-cluster#update-your-cluster-to-use-the-aks-gpu-image-preview
  * it adds a new capacity on the node of key "**nvidia.com/gpu**"
* PR that fix the example https://github.com/ryanzhang-oss/training-operator/pull/1/files
    * Fix the Dockerfile 
    * Play with the toy training python code
    * Play with the PyTorch yaml
* Tweak the model, build and push images, run the PytorchJob
    
## Issues
* Kubeflow Operator Issues
    * Doesn’t recreate pod if a pod is deleted
    * Master pod cannot stop before the work pod or the workers can’t continue
    * Why bother having a master definition at all when they are all the same
    * There can only be one master but I can set replica on Master.
    * Change the CRD doesn’t change the pod, once a pod is created, it’s immutable.
    * Overall, the opeartor is not a production ready one.
* Azure Issues
  * Autoscale not based on GPU
  * No easy way to see GPU usage on the portal
    * still haven't tried the AKS container insight.
 
  
## Run the LLaM on Pod
* follow https://github.com/Lightning-AI/lit-llama/blob/main/howto/finetune_lora.md running in the following pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    sidecar.istio.io/inject: "false"
  creationTimestamp: "2023-09-13T22:55:32Z"
  labels:
    training.kubeflow.org/replica-type: master
  name: pytorch-test
  namespace: default
spec:
  containers:
  - args:
    - infinity
    command:
    - sleep
    image: nvcr.io/nvidia/pytorch:23.08-py3
    imagePullPolicy: IfNotPresent
    name: pytorch
    ports:
    - containerPort: 23456
      name: pytorchjob-port
      protocol: TCP
    resources:
      limits:
        nvidia.com/gpu: "4"
      requests:
        nvidia.com/gpu: "4"
    securityContext:
      privileged: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
  nodeSelector:
    node.kubernetes.io:instance-type=standard_nc96ads_a100_v4
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: aks-a100-40057124-vmss000000
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoSchedule
    key: sku
    operator: Equal
    value: gpu
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  - effect: NoSchedule
    key: nvidia.com/gpu
    operator: Exists
```
One interesting thing I learned is that https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/#extendedresourcetoleration can auto tolerate pods
