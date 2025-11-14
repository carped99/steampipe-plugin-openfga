package openfga

import (
	"context"
	"fmt"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"sync"
	"time"
)

type Client struct {
	openfgav1.OpenFGAServiceClient
	conn    *grpc.ClientConn
	storeID string
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// connection 이름별로 Client 캐시
var (
	clientCache sync.Map // map[string]*Client
)

func getClient(ctx context.Context, d *plugin.QueryData) (*Client, error) {
	connName := d.Connection.Name
	if v, ok := clientCache.Load(connName); ok {
		return v.(*Client), nil
	}

	cfg, err := getConfig(d.Connection)
	if err != nil {
		return nil, err
	}

	client, err := newClient(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// LoadOrStore로 race 방지
	if actual, loaded := clientCache.LoadOrStore(connName, client); loaded {
		// 이미 다른 goroutine이 먼저 만든 경우, 우리가 만든 것은 닫아 줌
		_ = client.Close()
		return actual.(*Client), nil
	}

	return client, nil
}

func newClient(ctx context.Context, cfg *Config) (*Client, error) {
	// Extract endpoint from Endpoint
	var endpoint string
	if cfg.Endpoint != nil {
		endpoint = *cfg.Endpoint
	}
	if endpoint == "" {
		return nil, fmt.Errorf("api_url is required in connection config")
	}

	// Extract storeID
	var storeID string
	if cfg.StoreId != nil {
		storeID = *cfg.StoreId
	}

	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			// 큰 결과를 받을 수도 있으니 적당히 설정
			grpc.MaxCallRecvMsgSize(16 * 1024 * 1024),
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// For now, use insecure credentials (can be enhanced later with TLS support)
	dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Default timeout: 5 seconds
	timeout := 5 * time.Second
	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := grpc.DialContext(dialCtx, endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC endpoint %q: %w", endpoint, err)
	}

	// Create OpenFGA service client
	fgaServiceClient := openfgav1.NewOpenFGAServiceClient(conn)

	client := &Client{
		OpenFGAServiceClient: fgaServiceClient,
		conn:                 conn,
		storeID:              storeID,
	}
	return client, nil
}
