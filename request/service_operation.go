package request

import (
	"context"

	"github.com/go-camp/httpc"
)

type serviceNameKey struct{}
type operationNameKey struct{}

func GetServiceNameFromContext(ctx context.Context) (serviceName string) {
	ni := ctx.Value(serviceNameKey{})
	serviceName, _ = ni.(string)
	return
}

func GetOperationNameFromContext(ctx context.Context) (operationName string) {
	ni := ctx.Value(operationNameKey{})
	operationName, _ = ni.(string)
	return
}

func GetServiceNameFromMetadata(md httpc.Metadata) (serviceName string) {
	ni := md.Get(serviceNameKey{})
	serviceName, _ = ni.(string)
	return
}

func GetOperationNameFromMetadata(md httpc.Metadata) (operationName string) {
	ni := md.Get(operationNameKey{})
	operationName, _ = ni.(string)
	return
}

func GetServiceNameFromContextOrMetadata(ctx context.Context, md httpc.Metadata) (serviceName string) {
	serviceName = GetServiceNameFromContext(ctx)
	if serviceName == "" {
		serviceName = GetServiceNameFromMetadata(md)
	}
	return
}

func GetOperationNameFromContextOrMetadata(ctx context.Context, md httpc.Metadata) (operationName string) {
	operationName = GetOperationNameFromContext(ctx)
	if operationName == "" {
		operationName = GetOperationNameFromMetadata(md)
	}
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

func (ini WrapOperationErrorInitializer) initialize(ctx context.Context, input interface{}, initialize httpc.InitializeFunc) (
	output interface{}, md httpc.Metadata, err error,
) {
	output, md, err = initialize(ctx, input)
	if err != nil {
		serviceName := GetServiceNameFromContextOrMetadata(ctx, md)
		operationName := GetOperationNameFromContextOrMetadata(ctx, md)
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
