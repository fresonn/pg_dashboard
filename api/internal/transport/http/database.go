package http

import (
	"context"
	"dashboard/api/gen/openapi"
	"dashboard/api/internal/model/database"

	"github.com/go-playground/validator/v10"
)

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
