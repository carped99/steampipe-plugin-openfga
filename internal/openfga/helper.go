package openfga

import (
	"context"
	"fmt"
	openfgav1 "github.com/carped99/steampipe-plugin-openfga/internal/openfga/gen/openfga/v1"
	"github.com/google/martian/v3/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"sort"
)

// CreateOpenFGAClient creates and validates an OpenFGA service client
func CreateOpenFGAClient(cc grpc.ClientConnInterface, storeId, modelId string) (openfgav1.OpenFGAServiceClient, error) {
	serviceClient := openfgav1.NewOpenFGAServiceClient(cc)

	storeId, err := checkStore(serviceClient, storeId)
	if err != nil {
		return nil, err
	}

	log.Infof("OpenFGA Store: %s", storeId)
	fgaClient := NewOpenFGAServiceClient(cc, storeId, modelId)

	if err := validateModel(fgaClient, storeId, modelId); err != nil {
		return nil, err
	}

	return fgaClient, nil
}

func checkStore(client openfgav1.OpenFGAServiceClient, storeId string) (string, error) {
	if storeId == "" {
		log.Infof("Finding latest OpenFGA Store...")

		store, err := findLatestStore(client)
		if err != nil {
			return "", err
		}

		if store == nil {
			return "", fmt.Errorf("no active OpenFGA store found")
		}

		return store.GetId(), nil
	}

	store, err := client.GetStore(context.Background(), &openfgav1.GetStoreRequest{
		StoreId: storeId,
	})
	if err != nil {
		return "", fmt.Errorf("OpenFGA Store not found: %v, %w", storeId, err)
	}

	deletedAt := store.GetDeletedAt()
	if deletedAt != nil && !deletedAt.AsTime().IsZero() {
		return "", fmt.Errorf("OpenFGA Store is deleted: %s, [%s]", store.GetId(), store.GetName())
	}

	return store.GetId(), nil
}

// findLatestStore 가장 최근에 생성된 Store를 반환
func findLatestStore(client openfgav1.OpenFGAServiceClient) (*openfgav1.Store, error) {
	stores, err := findActiveStores(client)
	if err != nil {
		return nil, err
	}

	if len(stores) == 0 {
		return nil, nil
	}

	// 생성일 기준 정렬 (오름차순)
	sort.SliceStable(stores, func(i, j int) bool {
		return stores[i].GetCreatedAt().AsTime().Before(stores[j].GetCreatedAt().AsTime())
	})

	// 최신 항목 반환 (맨 마지막 요소)
	return stores[len(stores)-1], nil
}

// listAllStores 모든 Store를 조회
func listAllStores(client openfgav1.OpenFGAServiceClient) ([]*openfgav1.Store, error) {
	var (
		stores            []*openfgav1.Store
		continuationToken string
	)

	for {
		resp, err := client.ListStores(context.Background(), &openfgav1.ListStoresRequest{
			PageSize:          wrapperspb.Int32(100),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list stores: %w", err)
		}

		stores = append(stores, resp.GetStores()...)
		continuationToken = resp.GetContinuationToken()

		// 종료 조건: 더 이상 토큰이 없으면 끝
		if continuationToken == "" {
			break
		}
	}

	return stores, nil
}

// findActiveStores 활성화된 Store를 조회
func findActiveStores(client openfgav1.OpenFGAServiceClient) ([]*openfgav1.Store, error) {
	stores, err := listAllStores(client)
	if err != nil {
		return nil, err
	}

	// 삭제되지 않은 store만 필터링
	var activeStores []*openfgav1.Store
	for _, store := range stores {
		deletedAt := store.GetDeletedAt()
		if deletedAt == nil || deletedAt.AsTime().IsZero() {
			activeStores = append(activeStores, store)
		}
	}

	return activeStores, nil
}

// ValidateConnection validates the client connection
func ValidateConnection(ctx context.Context, conn grpc.ClientConnInterface) error {
	// 헬스 체크
	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return err
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("OpenFGA server is not serving: %s", resp.Status)
	}
	return nil
}

func validateModel(client openfgav1.OpenFGAServiceClient, storeId, modelId string) error {
	if modelId == "" {
		res, err := client.ReadAuthorizationModels(context.Background(), &openfgav1.ReadAuthorizationModelsRequest{
			StoreId: storeId,
		})
		if err != nil {
			return err
		}

		if len(res.GetAuthorizationModels()) == 0 {
			return fmt.Errorf("no OpenFGA AuthorizationModels found in store %s", storeId)
		}
	} else {
		_, err := client.ReadAuthorizationModel(context.Background(), &openfgav1.ReadAuthorizationModelRequest{
			StoreId: storeId,
			Id:      modelId,
		})

		if err != nil {
			return fmt.Errorf("OpenFGA AuthorizationModel not found: %s, %w", modelId, err)
		}
	}

	return nil
}
