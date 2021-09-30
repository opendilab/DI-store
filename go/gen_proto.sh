protoc --proto_path ./protos \
    --go_out=. \
    --go-grpc_out=. \
    ./protos/*.proto

rm -rf pb
mv di_store/pb ./
rmdir di_store

flatc --go protos/object.fbs