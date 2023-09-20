# Ray on AKS
This guide runs a sample Ray machine learning training workload with GPU on AKS.

The Ray documentation only had commands for deploying on GCP. The commands for deploying on AKS are below.
The original documentation that is based off is at https://docs.ray.io/en/latest/cluster/kubernetes/examples/gpu-training-example.html

A few modifications were needed to run the example on AKS:
- No explicit command is required to apply the daemonset when using VMs with GPUs
- When adding a nodepool with GPU, the aks-custom-header `UseGPUDedicatedVHD=true` should be set
- Added a toleration to the provided manifest file

## The end-to-end workflow
The following script summarizes the end-to-end workflow for GPU training. These instructions are for GCP, but a similar setup would work for any major cloud provider. The following script consists of:
- Step 1: Set up a Kubernetes cluster on AKS.
- Step 2: Deploy a Ray cluster on Kubernetes with the KubeRay operator.
- Step 3: Run the PyTorch image training benchmark.
```shell
# Step 0: Set up a resource group
az group create --name hackathonProject2 --location eastus

# Step 1: Set up a Kubernetes cluster on AKS
# Create a node-pool for a CPU-only head node
# standard_D8ds_v5 => 8 vCPU; 32 GB RAM
az aks create --resource-group hackathonProject2 --name gpu-cluster-1 --generate-ssh-keys --enable-cluster-autoscaler --node-count 1  --max-count 1 --min-count 1 --node-vm-size Standard_D8ds_v5

# Get credentials for the cluster
az aks get-credentials --resource-group hackathonProject2 --name gpu-cluster-1
 
# Create a node-pool for GPU. The node is for a GPU Ray worker node.
# standard_nc6s_v3 => 6 vCPU; 112 GB RAM
az aks nodepool add --resource-group hackathonProject2 --cluster-name gpu-cluster-1 --name gpunodepool --node-count 1 --node-vm-size standard_nc6s_v3 --node-taints sku=gpu:NoSchedule --aks-custom-headers UseGPUDedicatedVHD=true
  
# Step 2: Deploy a Ray cluster on Kubernetes with the KubeRay operator.

# Install both CRDs and KubeRay operator v0.6.0.
helm repo add kuberay https://ray-project.github.io/kuberay-helm/
helm install kuberay-operator kuberay/kuberay-operator --version 0.6.0
 
# Create a Ray cluster. Ensure that the end of the provided yaml file contains the following:
#         tolerations:
#       - key: "sku"
#         operator: "Equal"
#         value: "gpu"
#         effect: "NoSchedule"
kubectl apply -f ray-cluster.gpu-aks.yaml
 
# Set up port-forwarding
kubectl port-forward --address 0.0.0.0 services/raycluster-head-svc 8265:8265
 
# Step 3: Run the PyTorch image training benchmark.
# Install Ray if needed
pip3 install -U "ray[default]"
 
# Download the Python script
curl https://raw.githubusercontent.com/ray-project/ray/master/doc/source/cluster/doc_code/pytorch_training_e2e_submit.py -o pytorch_training_e2e_submit.py
 
# Submit the training job to your ray cluster
python3 pytorch_training_e2e_submit.py
 
# Use the following command to follow this Job's logs:
# Substitute the Ray Job's submission id.
ray job logs 'raysubmit_xxxxxxxxxxxxxxxx' --address http://127.0.0.1:8265 --follow
```
