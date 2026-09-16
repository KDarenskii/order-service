package cgrpc

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	ccatalog "github.com/KDarenskii/order-service/internal/app/client/catalog"
	"github.com/KDarenskii/order-service/internal/app/entity"
	catalogv1 "github.com/KDarenskii/order-service/internal/pkg/grpc/gen/catalog/v1"
)

type client struct {
	raw catalogv1.CatalogServiceClient
}

var _ ccatalog.Client = (*client)(nil)

func NewClient(address string) (ccatalog.Client, *grpc.ClientConn, error) {
	log.Info().
		Str("address", address).
		Msg("Initializing catalog-service gRPC client")

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create catalog-service grpc client: %w", err)
	}

	log.Info().Msg("Catalog-service gRPC client established")

	return &client{raw: catalogv1.NewCatalogServiceClient(conn)}, conn, nil
}

func (c *client) Ping(ctx context.Context) error {
	if _, err := c.raw.GetProduct(ctx, &catalogv1.GetProductRequest{}); err != nil {
		switch status.Code(err) {
		case codes.Unavailable, codes.DeadlineExceeded:
			return fmt.Errorf("ping catalog-service: %w", err)
		}
	}

	return nil
}

func (c *client) GetProduct(ctx context.Context, req *catalogv1.GetProductRequest) (*catalogv1.GetProductResponse, error) {
	res, err := c.raw.GetProduct(ctx, req)
	if err != nil {
		return nil, normalizeError(err)
	}

	return res, nil
}

func (c *client) GetProducts(ctx context.Context, req *catalogv1.GetProductsRequest) (*catalogv1.GetProductsResponse, error) {
	res, err := c.raw.GetProducts(ctx, req)
	if err != nil {
		return nil, normalizeError(err)
	}

	return res, nil
}

var grpcCodeToAppError = map[codes.Code]error{
	codes.NotFound:        entity.ErrNotFound,
	codes.InvalidArgument: entity.ErrIncorrectParameters,
	codes.AlreadyExists:   entity.ErrAlreadyExists,
}

func normalizeError(err error) error {
	st, ok := status.FromError(err)

	if !ok {
		return entity.ErrInternal
	}

	if appErr, ok := grpcCodeToAppError[st.Code()]; ok {
		return appErr
	}

	return entity.ErrInternal
}
