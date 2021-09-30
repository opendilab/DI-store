package util

import (
	"context"
	"errors"

	"di_store/tracing"

	otgrpc "github.com/opentracing-contrib/go-grpc"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	grpcServerOption []grpc.ServerOption
	grpcDialOption   = []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
)

func GrpcDialOption() []grpc.DialOption {
	return append(grpcDialOption,
		grpc.WithTimeout(CommonConfig.DialTimeout),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(CommonConfig.GrpcMaxCallRecvMsgSize)),
	)
}

func GrpcServerOption() []grpc.ServerOption {
	return grpcServerOption
}

func init() {
	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, CommonConfig.RequestTimeout)
		defer cancel()
		resp, err := handler(ctx, req)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				log.Warningf("method %q failed: %v", info.FullMethod, err)
			} else {
				log.Errorf("method %q failed: %+v", info.FullMethod, err)
			}
		} else {
			log.Debugf("method %q succeeded", info.FullMethod)
		}
		return resp, err
	}
	if tracing.Enabled {
		tracer := tracing.Tracer
		grpcServerOption = append(grpcServerOption,
			grpc.ChainUnaryInterceptor(errHandler, otgrpc.OpenTracingServerInterceptor(tracer)),
			grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)),
		)
		grpcDialOption = append(grpcDialOption,
			grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
			grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer)),
		)
	} else {
		grpcServerOption = append(grpcServerOption, grpc.UnaryInterceptor(errHandler))
	}

}
