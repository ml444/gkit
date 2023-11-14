package grpcx

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"

	"github.com/ml444/gkit/middleware/general"
)

// NewXDSConn new a connection of xDs
// Note: call `conn.Close()` when the server exits
func NewXDSConn(dsn string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	creds, err := xdscreds.NewClientCredentials(
		xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()},
	)
	if err != nil {
		return nil, err
	}
	conn, err = grpc.Dial(
		dsn,
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(general.ClientErrorInterceptor),
	)
	if err != nil {
		return nil, err
	}
	//NOTE: defer conn.Close()
	return conn, nil
}
