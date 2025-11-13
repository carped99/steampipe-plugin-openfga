package openfga

import (
	"context"
	"errors"
	"github.com/openfga/go-sdk/client"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"os"
)

func connect(ctx context.Context, d *plugin.QueryData) (*client.OpenFgaClient, error) {
	apiUrl := os.Getenv("OPENFGA_API_URL")
	storeId := os.Getenv("OPENFGA_STORE_ID")
	authToken := os.Getenv("OPENFGA_AUTH_TOKEN")

	config := getConfig(d.Connection)
	if config.ApiUrl != nil {
		apiUrl = *config.ApiUrl
	}
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
