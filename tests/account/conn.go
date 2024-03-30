package account

import (
	"errors"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"

	"github.com/ml444/gkit/log"
)

const EnvKeyAutoInjectConn = "AUTO_INJECT_CLIENT_CONN"
const EnvKeyAccountDNS = "SERVICE_ACCOUNT_DNS"

var dns string

func init() {
	var err error

	dns = os.Getenv(EnvKeyAccountDNS)
	if dns == "" {
		dns = "account.default.svc.cluster.local:5040"
		//dns = "xds:///uuid.default.svc.cluster.local:5040"
	}

	var isAutoInject bool
	autoInjectStr := os.Getenv(EnvKeyAutoInjectConn)
	if autoInjectStr != "" {
		isAutoInject, err = strconv.ParseBool(autoInjectStr)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
	}
	if isAutoInject {
		log.Infof("auto inject %s client conn", ClientName)
		conn, err := NewXDSConn()
		if err != nil {
			log.Fatal(err.Error())
			return
		}
		InjectConn(conn)
		go func() {
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, os.Kill)

			<-signalCh
			log.Warnf("exit: close %s client connect", ClientName)
			if err = conn.Close(); err != nil {
				log.Error(err.Error())
			}
		}()
	}
}

// NewXDSConn new a connection of xDs
// Note: call `conn.Close()` when the server exits
func NewXDSConn() (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	var err error
	creds, err := xdscreds.NewClientCredentials(
		xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()},
	)
	if err != nil {
		return nil, err
	}
	conn, err = grpc.Dial(dns, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}
	//NOTE: defer conn.Close()
	return conn, nil
}

type CliManager struct {
	conn    *grpc.ClientConn
	cli     AccountClient
	initErr error
}

func InjectConn(conn *grpc.ClientConn) {
	if conn == nil {
		return
	}
	cliMgr.conn = conn
	cliMgr.cli = NewAccountClient(conn)
}

func CloseConn() error {
	return cliMgr.conn.Close()
}

var cliMgr = CliManager{
	initErr: errors.New(
		"not yet initialized grpc client, " +
			"please call 'NewXDSConn()' and 'InjectConn()' on server, " +
			"and conn.Close() when the server exits"),
}
