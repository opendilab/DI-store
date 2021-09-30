package tracing

import (
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	opentracing "github.com/opentracing/opentracing-go"
	config "github.com/uber/jaeger-client-go/config"
)

var (
	Enabled bool = false
	Tracer  opentracing.Tracer
	gCloser io.Closer
)

func init() {
	endpoint := os.Getenv("JAEGER_ENDPOINT")
	if endpoint == "" {
		return
	}

	InitTracer("di_store")
	log.Infof("init tracer with endpoint %s", endpoint)
}

func Close() {
	gCloser.Close()
}

func InitTracer(service string) (opentracing.Tracer, io.Closer) {
	envConfig, err := config.FromEnv()
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}

	envConfig.ServiceName = service
	envConfig.Sampler.Type = "const"
	envConfig.Sampler.Param = 1
	envConfig.Reporter.LogSpans = true

	tracer, closer, err := envConfig.NewTracer(config.Logger(&jaegerLogger{}))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	opentracing.SetGlobalTracer(tracer)

	Enabled = true
	Tracer = tracer
	gCloser = closer

	return tracer, closer
}

type jaegerLogger struct{}

func (l *jaegerLogger) Error(msg string) {
	log.Error(msg)
}

func (l *jaegerLogger) Infof(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}
