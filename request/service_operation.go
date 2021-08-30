package request

import (
	"context"

	"github.com/go-camp/httpc"
)

type serviceNameKey struct{}
type operationNameKey struct{}

func GetServiceNameFromContext(ctx context.Context) (serviceName string, ok bool) {
	ni := ctx.Value(serviceNameKey{})
	serviceName, ok = ni.(string)
	return
}

func GetOperationNameFromContext(ctx context.Context) (operationName string, ok bool) {
	ni := ctx.Value(operationNameKey{})
	operationName, ok = ni.(string)
	return
}

func GetServiceNameFromMetadata(md httpc.Metadata) (serviceName string, ok bool) {
	ni, ok := md.Get(serviceNameKey{})
	if !ok {
		return "", ok
	}
	serviceName, ok = ni.(string)
	return
}

func GetOperationNameFromMetadata(md httpc.Metadata) (operationName string, ok bool) {
	ni, ok := md.Get(operationNameKey{})
	if !ok {
		return "", ok
	}
	operationName, ok = ni.(string)
	return
}

// ServiceOperationNameInitializer Add service and operatioin name to context and metadata.
type ServiceOperationNameInitializer struct {
	ServiceName   string
	OperationName string
}

func (ini ServiceOperationNameInitializer) Initializer(initialize httpc.InitializeFunc) httpc.InitializeFunc {
	return func(ctx context.Context, input interface{}) (output interface{}, md httpc.Metadata, err error) {
		return ini.initialize(ctx, input, initialize)
	}
}

func (ini ServiceOperationNameInitializer) initialize(ctx context.Context, input interface{}, initialize httpc.InitializeFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	ctx = context.WithValue(ctx, serviceNameKey{}, ini.ServiceName)
	ctx = context.WithValue(ctx, operationNameKey{}, ini.OperationName)

	output, md, err = initialize(ctx, input)

	md.Set(serviceNameKey{}, ini.ServiceName)
	md.Set(operationNameKey{}, ini.OperationName)

	return
}

// WrapOperationErrorInitializer wraps err as OperationError.
// WrapOperationErrorInitializer gets service and operation name from ServiceOperationNameInitializer.
type WrapOperationErrorInitializer struct{}

func (ini WrapOperationErrorInitializer) Initializer(initialize httpc.InitializeFunc) httpc.InitializeFunc {
	return func(ctx context.Context, input interface{}) (output interface{}, md httpc.Metadata, err error) {
		return ini.initialize(ctx, input, initialize)
	}
}

func getServiceNameFromContextOrMetadata(ctx context.Context, md httpc.Metadata) (serviceName string, ok bool) {
	serviceName, ok = GetServiceNameFromContext(ctx)
	if ok {
		return
	}
	return GetServiceNameFromMetadata(md)
}

func getOperationNameFromContextOrMetadata(ctx context.Context, md httpc.Metadata) (operationName string, ok bool) {
	operationName, ok = GetOperationNameFromContext(ctx)
	if ok {
		return
	}
	return GetOperationNameFromMetadata(md)
}

func (ini WrapOperationErrorInitializer) initialize(ctx context.Context, input interface{}, initialize httpc.InitializeFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	output, md, err = initialize(ctx, input)
	if err != nil {
		serviceName, _ := getServiceNameFromContextOrMetadata(ctx, md)
		operationName, _ := getOperationNameFromContextOrMetadata(ctx, md)
		if serviceName != "" || operationName != "" {
			err = &httpc.OperationError{
				Service:   serviceName,
				Operation: operationName,
				Err:       err,
			}
		}
	}
	return
}
