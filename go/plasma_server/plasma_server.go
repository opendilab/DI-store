package plasma_server

import (
	"context"
	"di_store/util"
	_ "embed"
	"strconv"
)

//plasma-store-server

//go:embed plasma-store-server
var embeddedBinary []byte

func RunPlasmaStoreServer(ctx context.Context, memoryByte int, socketPath string) (context.Context, error) {
	p, err := util.CreateMemFileFromBytes(embeddedBinary)
	if err != nil {
		return nil, err
	}
	return util.RunCmd(ctx, p, "-m", strconv.Itoa(memoryByte), "-s", socketPath), nil
}
