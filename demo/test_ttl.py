import time

from di_store import Client

client = Client('../conf/di_store.yaml')
ref1 = client.put(b'Hello world.', ttl=10)
ref2 = client.put(b'no ttl')
ref3 = client.put(b'sdsdsd', ttl=2)
ref4 = client.put(b'Hello world.', ttl=2)
print(client.getExpireSet())  # 3 objects in expire_set
time.sleep(3)
ref5 = client.put(b'Hello world.', ttl=20)
print(client.getExpireSet())  # 3 objects in expire_set

data = client.get(ref1)
print('data:', data)
# client.delete(ref3)
data = client.get(ref3)
print('data:', data)
