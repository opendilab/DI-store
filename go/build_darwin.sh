if [ ! -d release ]; then
  mkdir -p release
fi

echo "building storage_server_darwin"
go build -a -o release/storage_server_darwin storage_server/main/storage_server.go
go build -race -a -o release/storage_server_darwin.race storage_server/main/storage_server.go

echo "building node_tracker_darwin"
go build -a -o release/node_tracker_darwin node_tracker/main/node_tracker.go
go build -race -a -o release/node_tracker_darwin.race node_tracker/main/node_tracker.go
