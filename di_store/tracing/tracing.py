import os
from jaeger_client import Config
from urllib.parse import urlparse
from functools import partial, wraps
import inspect

from grpc_opentracing import open_tracing_client_interceptor
from grpc_opentracing.grpcext import intercept_channel

tracer = None
tracer_interceptor = None


def init_tracer(service):
    config = Config(
        config={  # usually read from some yaml config
            'sampler': {
                'type': 'const',
                'param': 1,
                # 'type': 'probabilistic',
                # 'param': 0.005,

            },
            'reporter_batch_size': 100,
            'reporter_queue_size': 100,
            # 'reporter_flush_interval': 1000,
            # 'sampling_refresh_interval': 1000
        },
        service_name=service,
    )
    # this call also sets opentracing.tracer
    return config.initialize_tracer()


endpoint = os.getenv('JAEGER_ENDPOINT')
if endpoint:
    result = urlparse(endpoint)
    if not result.hostname:
        raise ValueError(
            f'can not get hostname from JAEGER_ENDPOINT: {endpoint}')
    os.environ['JAEGER_AGENT_HOST'] = result.hostname

if os.getenv('JAEGER_AGENT_HOST'):
    print('JAEGER_AGENT_HOST:', os.getenv('JAEGER_AGENT_HOST'))
    tracer = init_tracer('petrel_rl_store_client')
    tracer_interceptor = open_tracing_client_interceptor(tracer)


def wrap_channel(channel):
    if tracer_interceptor:
        channel = intercept_channel(channel, tracer_interceptor)

    return channel


def trace_callable(target, *, operation_name=None, span_name=None):
    if getattr(target, '__traced', False):
        return target

    if operation_name is None:
        operation_name = target.__qualname__

    @wraps(target)
    def wrapped(*args, **kwargs):

        with tracer.start_active_span(operation_name) as scope:
            if span_name:
                kwargs[span_name] = scope.span
            return target(*args, **kwargs)

    setattr(wrapped, '__traced', True)
    return wrapped


def trace_class(cls, *, span_name=None):
    for attr, val in cls.__dict__.items():
        if not attr.startswith('__') and callable(val):
            setattr(cls, attr, trace(val, span_name=span_name))
    return cls


def trace(target=None, *, operation_name=None, span_name=None):
    if target is None:
        return partial(trace, operation_name=operation_name, span_name=span_name)

    if tracer is None:
        return target

    if inspect.isclass(target):
        if operation_name is not None:
            raise ValueError(
                'operation_name is not supported for class decoration')
        return trace_class(target, span_name=span_name)
    elif callable(target):
        return trace_callable(target, operation_name=operation_name, span_name=span_name)
    else:
        raise ValueError(f'can not set decorator on type {type(target)}')
