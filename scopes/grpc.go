package scopes

import (
	"context"
	"fmt"
	"strings"

	"github.com/dapperlabs/protoc-gen-go-grpc-scopes/scopesproto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ScopeValidator func(ctx context.Context, scopes []string) error

func getMethodDescriptor(info *grpc.UnaryServerInfo) (protoreflect.MethodDescriptor, error) {
	// /package.Service/Method
	fullMethodSplit := strings.Split(info.FullMethod, "/")

	if len(fullMethodSplit) != 3 {
		return nil, fmt.Errorf("invalid full method: %s", info.FullMethod)
	}

	svcFullName := protoreflect.FullName(fullMethodSplit[1])

	descriptorByService, err := protoregistry.GlobalFiles.FindDescriptorByName(svcFullName)
	if err != nil {
		return nil, fmt.Errorf("failed to find service descriptor: %s", err)
	}

	svc, ok := descriptorByService.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("descriptor is not of a service: %s", descriptorByService.FullName())
	}

	serviceMethod := protoreflect.FullName(info.FullMethod).Name()

	// Service/Method
	serviceMethodSplit := strings.Split(string(serviceMethod), "/")
	if len(serviceMethodSplit) != 2 {
		return nil, fmt.Errorf("invalid service method: %s", serviceMethod)
	}

	methodName := protoreflect.Name(serviceMethodSplit[1])

	method := svc.Methods().ByName(methodName)
	if method == nil {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	return method, nil
}

// ScopeValidationInterceptor validates that the request has the required scopes
func ScopeValidationInterceptor(v ScopeValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		method, err := getMethodDescriptor(info)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get method descriptor: %s", err)
		}

		methodOptionsPB := method.Options().(*descriptorpb.MethodOptions)

		methodExt := proto.GetExtension(methodOptionsPB, scopesproto.E_RequiredMethodScopes)
		if methodExt != nil {
			scopesExt := methodExt.(*scopesproto.RequiredScopesOption)

			if err := v(ctx, scopesExt.GetScopes()); err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}
