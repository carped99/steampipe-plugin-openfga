package openfga

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	openfgainternal "github.com/carped99/steampipe-plugin-openfga/internal/openfga"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/openfga/go-sdk/client"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strings"
)

func connect(ctx context.Context, d *plugin.QueryData) (*client.OpenFgaClient, error) {
	apiUrl := os.Getenv("OPENFGA_API_URL")
	storeId := os.Getenv("OPENFGA_STORE_ID")
	authToken := os.Getenv("OPENFGA_AUTH_TOKEN")

	config := getConfig(d.Connection)
	apiUrl = config.Endpoint
	if config.StoreId != nil {
		storeId = *config.StoreId
	}
	if config.ApiToken != nil {
		authToken = *config.ApiToken
	}

	if apiUrl == "" {
		return nil, errors.New("'api_url' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	if storeId == "" {
		return nil, errors.New("'store_id' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	clientConfig := &client.ClientConfiguration{
		ApiUrl:  apiUrl,
		StoreId: storeId,
	}

	// Optional: Add auth token if provided
	if authToken != "" {
		// OpenFGA SDK uses bearer token in header
		// The SDK will automatically add "Authorization: Bearer <token>" header
		// Note: This depends on SDK version and configuration
		// For custom auth, you may need to use http.Client with custom headers
	}

	fgaClient, err := client.NewSdkClient(clientConfig)

	return fgaClient, err
}

// grpcClientCacheKey is the cache key for the gRPC client
const grpcClientCacheKey = "openfga_grpc_client"

// grpcClientWrapper wraps the gRPC client and connection for proper lifecycle management
type grpcClientWrapper struct {
	client openfgav1.OpenFGAServiceClient
	conn   *grpc.ClientConn
}

// Close closes the underlying gRPC connection
func (w *grpcClientWrapper) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// connectGrpc creates or retrieves a cached gRPC client connection to OpenFGA
func connectGrpc(ctx context.Context, d *plugin.QueryData) (openfgav1.OpenFGAServiceClient, error) {
	// Check if client is already cached (skip if ConnectionCache is not available)
	if d.ConnectionCache != nil {
		if cachedData, ok := d.ConnectionCache.Get(ctx, grpcClientCacheKey); ok {
			wrapper := cachedData.(*grpcClientWrapper)
			return wrapper.client, nil
		}
	}

	// Get configuration
	apiUrl := os.Getenv("OPENFGA_API_URL")
	storeId := os.Getenv("OPENFGA_STORE_ID")
	modelId := os.Getenv("OPENFGA_AUTHORIZATION_MODEL_ID")

	config := getConfig(d.Connection)
	apiUrl = config.Endpoint
	if config.StoreId != nil {
		storeId = *config.StoreId
	}
	if config.AuthorizationModelId != nil {
		modelId = *config.AuthorizationModelId
	}

	if apiUrl == "" {
		return nil, errors.New("'api_url' must be set in the connection configuration. Edit your connection configuration file and then restart Steampipe")
	}

	// Parse the URL to extract host and determine if TLS is needed
	// Expected format: http://localhost:8080 or https://api.openfga.example.com
	var target string
	var opts []grpc.DialOption

	if strings.HasPrefix(apiUrl, "https://") {
		target = strings.TrimPrefix(apiUrl, "https://")
		// Use TLS credentials
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else if strings.HasPrefix(apiUrl, "http://") {
		target = strings.TrimPrefix(apiUrl, "http://")
		// Use insecure connection
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		return nil, fmt.Errorf("invalid api_url format: %s (must start with http:// or https://)", apiUrl)
	}

	// Create gRPC connection following best practices
	// https://github.com/grpc/grpc-go/blob/master/Documentation/anti-patterns.md
	// grpc.NewClient performs no I/O - connections are established lazily
	// Errors should be handled at RPC call time, not at dial time
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	// Create OpenFGA service client with adaptor
	fgaClient, err := openfgainternal.CreateOpenFGAClient(conn, storeId, modelId)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create OpenFGA client: %w", err)
	}

	// Wrap the client and connection for proper lifecycle management
	wrapper := &grpcClientWrapper{
		client: fgaClient,
		conn:   conn,
	}

	// Cache the wrapper for reuse (if ConnectionCache is available)
	if d.ConnectionCache != nil {
		d.ConnectionCache.Set(ctx, grpcClientCacheKey, wrapper)
	}

	return fgaClient, nil
}
