package util

import (
	"context"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type CmdOutputWriter struct {
	buff        []byte
	disableBuff bool
}

func (writer *CmdOutputWriter) Write(p []byte) (n int, err error) {
	if !writer.disableBuff {
		writer.buff = append(writer.buff, p...)
	}
	log.Debug(string(p))
	return len(p), nil
}

func RunCmd(ctx context.Context, name string, arg ...string) error {
	ctx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(ctx, name, arg...)
	writer := &CmdOutputWriter{}
	cmd.Stdout = writer
	cmd.Stderr = writer
	errChan := make(chan error)

	go func() {
		defer cancel()
		err := cmd.Run()
		if err != nil {
			err = errors.Wrapf(err, "util.RunCmd: %+v", cmd)
			errChan <- err
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Infof("cancel command: %v", cmd)
		cancel()
	}()

	select {
	case <-time.After(time.Second):
		writer.disableBuff = true
		writer.buff = nil
		return nil
	case err := <-errChan:
		return errors.Wrapf(err, string(writer.buff))
	}
}
