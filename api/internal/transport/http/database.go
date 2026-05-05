package http

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/model/database"
	databaseService "dashboard/api/internal/service/database"
	"errors"

	"github.com/go-playground/validator/v10"
)

func (h *Handler) Database(ctx context.Context, req openapi.DatabaseRequestObject) (openapi.DatabaseResponseObject, error) {

	db, err := h.database.Database(ctx, req.DatabaseId)
	if err != nil {
		if errors.Is(err, databaseService.ErrNotFound) {
			return openapi.Database404JSONResponse{
				Message: err.Error(),
			}, nil
		}

		return openapi.Database400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.Database200JSONResponse(db), nil
}

func (h *Handler) DatabasesDetailed(ctx context.Context, request openapi.DatabasesDetailedRequestObject) (openapi.DatabasesDetailedResponseObject, error) {

	params := request.Params

	if err := validator.New(validator.WithRequiredStructEnabled()).Struct(&params); err != nil {
		return openapi.DatabasesDetailed422JSONResponse{
			Message: "Request validation failed",
			Reason:  err.Error(),
		}, nil
	}

	var filter database.DatabasesFilter

	if params.Sort != nil {
		filter.Sort = string(*params.Sort)
	}

	if params.Order != nil {
		filter.Order = string(*params.Order)
	}

	databases, err := h.database.DatabasesDetailed(ctx, filter)
	if err != nil {
		return openapi.DatabasesDetailed400JSONResponse{
			Message: err.Error(),
		}, nil
	}

	return openapi.DatabasesDetailed200JSONResponse(databases), nil
}
