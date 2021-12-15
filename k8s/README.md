# Running DI-store on Kubernetes


### Build the DI-store docker image


```bash
docker build -t <di-store-image-tag> <path-to-project-root>
```

### Generate k8s yaml files


```bash
APP_ID=<app-id>
IMAGE_TAG=<di-store-image-tag>
sed "s#__APP_ID__#$APP_ID#g; s#__IMAGE_TAG__#$IMAGE_TAG#g" di_store_k8s_template.yaml > di_store_k8s.yaml
sed "s#__APP_ID__#$APP_ID#g" di_store_client_template.yaml > di_store_client.yaml
```

The `<app-id>` should be unique for each DI-store engine instance.
Please increase the memory limit if necessary in the `di_store_k8s_template.yaml` file.

### Create an instance of DI-store engine


```bash
kubectl create -f di_store_k8s.yaml
```
Specify the namespace with `-n <namespace>` if necessary.

### Configuration for Pods using DI-store client

Add lines below (marked with `###`) to the k8s yaml file.


```yaml
...
spec:
  ...
  shareProcessNamespace: true       ###
  ...
  volumes:
  - name: shared-memory             ###
    hostPath:                       ###
      path: /dev/shm                ###
  ...
  containers:
  - name: your-container
    ...
    env:
    - name: DI_STORE_NODE_NAME      ###
      valueFrom:                    ###
        fieldRef:                   ###
          fieldPath: spec.nodeName  ###
    ...
    volumeMounts:
      - name: shared-memory         ###
        mountPath: /dev/shm         ###
...
```

The DI-store should be installed in the container's python environment, and the `di_store_client.yaml` file generated from the second step needs to be copied to the container.

A demo of using DI-store client:
```python
from di_store import Client
client = Client(<path to di_store_client.yaml>)
...
ref = client.put(data)
data2 = client.get(ref)
...
```

### Data prefetching

Register one or more clients with group name `<group_name>`
```python
client.register_group(<group_name>)
```

put data on DI-store with `prefetch_group` parameter
```python
ref = client.put(data, prefetch_group=<group_name>)
```

Once the `put` method returns, clients registered with the corresponding group name begin to prefetch data in the background. 

### Shutdown the DI-store engine instance


```bash
APP_ID=<app-id>
kubectl delete service node-tracker-$APP_ID
kubectl delete pod node-tracker-$APP_ID
kubectl delete daemonset storage-server-daemonset-$APP_ID
```

### Note

1. It's recommended to start an individual instance of DI-store engine for each application and shut down the instance when the application exits. The instance is able to keep running for a long time and serve multiple applications as well, which requires more CPU and memory resources. Make sure to delete objects that are no longer used to avoid memory leaks for long running.
1. Whenever an instance launches (via `kubectl create -f di_store_k8s.yaml`), a pod is created for running `etcd_server` and `node_tracker` processes. At the same time, a pod running `storage_server` is started on each k8s node respectively (by default). In order to constrain the `storage_server` processes running on a particular set of nodes, `node selector` or `node affinity` should be added to the configuration of `storage-server-daemonset` in the k8s yaml file. More information about running pods on selected nodes can be found [here](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/#running-pods-on-select-nodes). Also notice that, for any node having pod running DI-store client, the pod running `storage_server` should be started at that node as well.
1. For any pod running DI-store client, the specific `di_store_client.yaml` file generated together with the `di_store_k8s.yaml` should be used to initialize the client, so that the client will connect to the corresponding DI-store engine instance.