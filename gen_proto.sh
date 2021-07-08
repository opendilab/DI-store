python -m grpc_tools.protoc -I ./protos \
--python_out=./di_store/storage \
--grpc_python_out=./di_store/storage \
./protos/storage_server.proto

python -m grpc_tools.protoc -I ./protos \
--python_out=./di_store/node_tracker \
--grpc_python_out=./di_store/node_tracker \
./protos/node_tracker.proto

sed -i .bak  '/import .* as / s/./from . &/' ./di_store/**/*_grpc.py