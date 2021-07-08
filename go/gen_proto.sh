protoc --proto_path ./protos \
    --go_out=. \
    --go-grpc_out=. \
    ./protos/*.proto