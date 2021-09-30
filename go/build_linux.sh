if [ ! -d release ]; then
  mkdir -p release
fi

scl enable devtoolset-7 - << \EOF
echo "gcc: `which gcc`"
go env -w GO111MODULE=on

echo "building storage_server_linux"
go build -a -o release/storage_server_linux storage_server/main/storage_server.go
go build -race -a -o release/storage_server_linux.race storage_server/main/storage_server.go

echo "building node_tracker_linux"
go build -a -o release/node_tracker_linux node_tracker/main/node_tracker.go
go build -race -a -o release/node_tracker_linux.race node_tracker/main/node_tracker.go
EOF
