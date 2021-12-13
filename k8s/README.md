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

### Set `shareProcessNamespace` to `true`, and define the `DI_STORE_NODE_NAME` environment variable for your container


```yaml
...
spec:
  ...
  shareProcessNamespace: true
  ...
  containers:
  - name: your-container
    ...
    env:
    - name: DI_STORE_NODE_NAME
      valueFrom:
        fieldRef:
          fieldPath: spec.nodeName
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