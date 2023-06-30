package scopes

import (
	"context"

	"google.golang.org/grpc"
)

type ScopeValidator func(ctx context.Context, scopes []string) error

// ScopeValidationInterceptor validates that the request has the required scopes
func ScopeValidationInterceptor(v ScopeValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if scopeFetcher, ok := req.(HasScopeRequirements); ok {
			if err := v(ctx, scopeFetcher.RequiredScopes()); err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}
