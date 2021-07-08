## Introduction
Decision AI Store

<div align="center">
  <a href="https://github.com/opendilab/DI-store"><img width="700px" height="auto" src="di_store.svg"></a>
</div>


## Installation

##### Prerequisites
- Linux

- Python >= 3.6


```bash
pip install .
```

## Quick Start

##### Start Etcd Server

```bash
di_store etcd_server ./conf/di_store.yaml
```

##### Start Node Tracker
```bash
di_store node_tracker ./conf/di_store.yaml
```

##### Start Storage Server

```bash
di_store storage_server ./conf/di_store.yaml
```

##### Start Storage Client

```python
from di_store import Client
client = Client('./conf/di_store.yaml')
ref = client.put(b'Hello world.')
data = client.get(ref)
print('data:', data)
client.delete(ref)
```

## License
DI-store released under the Apache 2.0 license.
