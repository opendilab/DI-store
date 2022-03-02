from di_store import Client

client = Client('../conf/di_store.yaml')
ref = client.put(b'Hello world.')
data = client.get(ref)
print('data:', data)
client.delete(ref)
