# sample
[sample job](https://github.com/kubeflow/training-operator/blob/master/examples/mpi/tensorflow-mnist.yaml)

## issue
infiniband nics are not available because pod doesn't have ib device in pod network namespace. It has a huge impact on performance. 

