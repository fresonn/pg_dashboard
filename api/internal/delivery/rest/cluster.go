package rest

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/scopes/cluster/entities"
	"errors"

	"github.com/go-playground/validator/v10"
)

func (h *Handler) GetStatus(ctx context.Context, request openapi.GetStatusRequestObject) (openapi.GetStatusResponseObject, error) {

	status := h.cluster.PostgresStatus(ctx)

	resp := openapi.GetStatusResponse{
		PostgresConnection: openapi.ConnectionStatus(status),
	}

	return openapi.GetStatus200JSONResponse(resp), nil
}

func (h *Handler) ClusterConnect(ctx context.Context, request openapi.ClusterConnectRequestObject) (openapi.ClusterConnectResponseObject, error) {

	err := h.cluster.Connect(ctx, entities.AuthData(*request.Body))
	if err != nil {

		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return openapi.ClusterConnect422JSONResponse{
				Message: "Request validation failed",
				Reason:  err.Error(),
			}, nil
		}

		return openapi.ClusterConnect400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.ClusterConnect200JSONResponse{
		PostgresConnection: openapi.PgConnectionStatusConnected,
	}, nil
}

func (h *Handler) ClusterDisconnect(ctx context.Context, _request openapi.ClusterDisconnectRequestObject) (openapi.ClusterDisconnectResponseObject, error) {

	err := h.cluster.Disconnect(ctx)
	if err != nil {
		return openapi.ClusterDisconnect400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.ClusterDisconnect200JSONResponse{
		PostgresConnection: openapi.PgConnectionStatusDisconnected,
	}, nil
}

func (h *Handler) PostgresVersion(ctx context.Context, request openapi.PostgresVersionRequestObject) (openapi.PostgresVersionResponseObject, error) {

	version, err := h.cluster.Version(ctx)
	if err != nil {
		return openapi.PostgresVersion400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.PostgresVersion200JSONResponse(version), nil
}

func (h *Handler) PostgresUptime(ctx context.Context, request openapi.PostgresUptimeRequestObject) (openapi.PostgresUptimeResponseObject, error) {

	uptime, err := h.cluster.Uptime(ctx)
	if err != nil {
		return openapi.PostgresUptime400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.PostgresUptime200JSONResponse(uptime), nil
}

func (h *Handler) PostmasterSettings(ctx context.Context, request openapi.PostmasterSettingsRequestObject) (openapi.PostmasterSettingsResponseObject, error) {

	settings, err := h.cluster.PostmasterSettings(ctx)
	if err != nil {
		return openapi.PostmasterSettings400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.PostmasterSettings200JSONResponse(settings), nil
}

func (h *Handler) DatabasesDetailed(ctx context.Context, request openapi.DatabasesDetailedRequestObject) (openapi.DatabasesDetailedResponseObject, error) {

	params := request.Params

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(&params); err != nil {
		return openapi.DatabasesDetailed422JSONResponse{
			Message: "Request validation failed",
			Reason:  err.Error(),
		}, nil
	}

	var filter entities.DatabasesFilter

	if params.Sort != nil {
		filter.Sort = string(*params.Sort)
	}

	if params.Order != nil {
		filter.Order = string(*params.Order)
	}

	databases, err := h.cluster.DatabasesDetailed(ctx, filter)
	if err != nil {
		return openapi.DatabasesDetailed400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.DatabasesDetailed200JSONResponse(databases), nil
}
