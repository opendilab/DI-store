package plasma_server

import (
	"context"
	"di_store/util"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

func RunPlasmaStoreServer(ctx context.Context, memoryByte int, socketPath string) error {
	execPath := os.Getenv("PLASMA_STORE_SERVER_EXEC")
	if execPath == "" {
		return errors.Errorf("environment variable PLASMA_STORE_SERVER_EXEC not found")
	}

	return util.RunCmd(ctx, execPath, "-m", strconv.Itoa(memoryByte), "-s", socketPath)
}
