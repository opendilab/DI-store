if [ ! -d release ]; then
  mkdir -p release
fi

scl enable devtoolset-7 - << \EOF
echo "gcc: `which gcc`"
go env -w GO111MODULE=on

echo "building storage_server"
go build -a -o release/storage_server storage_server/main/storage_server.go
go build -race -a -o release/storage_server.race storage_server/main/storage_server.go

echo "building node_tracker"
go build -a -o release/node_tracker node_tracker/main/node_tracker.go
go build -race -a -o release/node_tracker.race node_tracker/main/node_tracker.go
EOF
