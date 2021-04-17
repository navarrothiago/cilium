package endpoint

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/spiffe/go-spiffe/v2/proto/spiffe/workload"

	"google.golang.org/grpc"
	//"google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc/metadata"
)

/* spire logic starts */
const (
	SpireSocketPath = "/run/spire/sockets/agent.sock"
)

var errContainerNotFound = errors.New("container not found")

func dialer(ctx context.Context, addr string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "unix", addr)
}

func UDSDial(socketPath string) (*grpc.ClientConn, error) {
	return grpc.Dial(socketPath,
		grpc.WithInsecure(),
		grpc.WithContextDialer(dialer),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
		/*grpc.WithReturnConnectionError()*/)
}

func NewWorkloadClient(conn *grpc.ClientConn) workload.SpiffeWorkloadAPIClient {
	return workload.NewSpiffeWorkloadAPIClient(conn)
}

func getClient() (workload.SpiffeWorkloadAPIClient, error) {
	conn, err := UDSDial(SpireSocketPath)
	if err != nil {
		return nil, fmt.Errorf("dialing: %v", err)
	}

	client := NewWorkloadClient(conn)
	if client == nil {
		fmt.Printf("error creating entry client")
	}

	return client, nil
}

func getStream(podUuid string) (workload.SpiffeWorkloadAPI_FetchX509SVIDClient, error) {
	client, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("error getting client: %w", err)
	}

	header := metadata.Pairs("workload.spiffe.io", "true")
	ctx := metadata.NewOutgoingContext(context.Background(), header)

	req := workload.X509SVIDRequest{
		Credentials: &workload.X509SVIDRequest_PodUuid{
			PodUuid: podUuid,
		},
	}

	return client.FetchX509SVID(ctx, &req)
}
