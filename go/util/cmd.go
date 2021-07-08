package util

import (
	"context"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type CmdOutputWriter struct{}

func (writer *CmdOutputWriter) Write(p []byte) (n int, err error) {
	log.Debug(string(p))
	return len(p), nil
}

func RunCmd(ctx context.Context, name string, arg ...string) context.Context {
	cmdCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, name, arg...)
	writer := &CmdOutputWriter{}
	cmd.Stdout = writer
	cmd.Stderr = writer
	go func() {
		defer cancel()
		err := cmd.Run()
		if err != nil {
			log.Error(err)
		}
	}()

	// go func() {
	// 	<-ctx.Done()
	// 	cancel()
	// }()
	return cmdCtx
}
