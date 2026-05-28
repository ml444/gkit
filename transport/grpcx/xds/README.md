# grpcx/xds Usage

This package provides a thin wrapper for gRPC xDS client/server setup in gkit.

## Prerequisites

- Your environment has a reachable xDS control plane.
- `GRPC_XDS_BOOTSTRAP` points to a valid bootstrap JSON file.
- Target uses xDS scheme, for example: `xds:///listener-name`.

Example bootstrap:

```bash
export GRPC_XDS_BOOTSTRAP=/path/to/bootstrap.json
```

## Client

Use `xds.NewClient` to create a `*grpc.ClientConn`.

```go
package main

import (
	"log"

	"google.golang.org/grpc"

	gkitxds "github.com/ml444/gkit/transport/grpcx/xds"
)

func main() {
	conn, err := gkitxds.NewClient(
		"xds:///user-service-listener",
		gkitxds.WithDialOptions(
			grpc.WithBlock(),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// use generated gRPC client stubs with conn
}
```

Notes:

- `NewClient` configures xDS credentials with insecure fallback.
- `general.ClientErrorInterceptor` is included by default.

## Server

Use `xds.NewGRPCServer` to create an xDS-backed gRPC server.

```go
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	gkitxds "github.com/ml444/gkit/transport/grpcx/xds"
)

func main() {
	srv, err := gkitxds.NewGRPCServer(
		grpc.MaxRecvMsgSize(8<<20),
	)
	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", ":5040")
	if err != nil {
		log.Fatal(err)
	}

	// register pb services to srv here
	log.Fatal(srv.Serve(lis))
}
```

## Choosing discovery mechanism

Do not mix `xds:///` and `discovery:///` for the same `ClientConn`.

- Use `xds:///...` for service mesh / xDS control-plane driven routing.
- Use `discovery:///...` with `transport/grpcx` + `discovery` module for registry-based discovery.
