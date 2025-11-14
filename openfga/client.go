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

// clearClientCache clears all cached clients (for testing)
func clearClientCache() {
	clientCache.Range(func(key, value interface{}) bool {
		clientCache.Delete(key)
		return true
	})
}

func getClient(ctx context.Context, d *plugin.QueryData) (*Client, error) {
	connName := d.Connection.Name
	if v, ok := clientCache.Load(connName); ok {
		return v.(*Client), nil
	}

	cfg := getConfig(d.Connection)

	client, err := NewClient(ctx, cfg)
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

// NewClient creates a new OpenFGA client with the given configuration
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// Validate endpoint
	if cfg.Endpoint == "" {
		return nil, fmt.Errorf("endpoint is required in connection config")
	}

	// Extract storeID
	var storeID string
	if cfg.StoreId != nil {
		storeID = *cfg.StoreId
	}

	// Configure dial options following gRPC best practices
	// https://github.com/grpc/grpc-go/blob/master/Documentation/anti-patterns.md
	dialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(16 * 1024 * 1024), // 16MB max receive message size
		),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                60 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// Configure TLS/credentials
	useTLS := false
	if cfg.UseTLS != nil {
		useTLS = *cfg.UseTLS
	}

	if useTLS {
		// TLS configuration will be added here in the future
		// For now, use insecure credentials
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// Use insecure credentials for non-TLS connections
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Use grpc.NewClient (recommended since v1.63.0)
	// This performs NO I/O during construction - connections are established lazily
	// Errors should be handled at RPC call time, not at dial time
	conn, err := grpc.NewClient(cfg.Endpoint, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client for endpoint %q: %w", cfg.Endpoint, err)
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
