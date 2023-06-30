package scopes_test

import (
	"context"
	"net"
	"testing"

	"github.com/dapperlabs/proto-gen-go-grpc-scopes/scopes"
	"github.com/dapperlabs/proto-gen-go-grpc-scopes/test/testgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestScopeValidationInterceptor(t *testing.T) {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(scopes.ScopeValidationInterceptor(func(ctx context.Context, scopes []string) error {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return status.Error(codes.Unauthenticated, "missing metadata")
			}

			authScope := md.Get("authorization_scope")
			if len(authScope) != 1 {
				return status.Error(codes.Unauthenticated, "missing authorization_scope")
			}
			providedScope := authScope[0]
			for _, allowedScope := range scopes {
				if providedScope == allowedScope {
					return nil
				}
			}
			return status.Errorf(codes.PermissionDenied, "missing scope: %s", scopes)
		})),
	)
	testgen.RegisterPingPongServer(server, &ScopeValidatorServer{})

	lis, err := net.Listen("tcp", ":8080")
	require.NoError(t, err)

	go server.Serve(lis)

	conn, err := grpc.Dial(":8080", grpc.WithInsecure())
	require.NoError(t, err)
	defer conn.Close()

	client := testgen.NewPingPongClient(conn)

	header := metadata.New(map[string]string{"authorization_scope": "scope1"})
	ctx := metadata.NewOutgoingContext(context.Background(), header)

	res, err := client.Ping(ctx, &testgen.PingRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "pong", res.GetPong())

	res, err = client.Ping(context.Background(), &testgen.PingRequest{})
	assert.Equal(t, codes.Unauthenticated, status.Code(err))

	header = metadata.New(map[string]string{"authorization_scope": "scope3"})
	ctx = metadata.NewOutgoingContext(context.Background(), header)
	res, err = client.Ping(ctx, &testgen.PingRequest{})
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

type ScopeValidatorServer struct {
}

func (s ScopeValidatorServer) Ping(ctx context.Context, request *testgen.PingRequest) (*testgen.PingResponse, error) {
	return &testgen.PingResponse{Pong: "pong"}, nil
}

var _ testgen.PingPongServer = (*ScopeValidatorServer)(nil)
