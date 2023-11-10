package refresh

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/server/handlers"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
)

type Request struct {
	GUID         string `json:"id" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

type HandlerRefresh struct {
	log         *logging.Logger
	userService *services.UserService
}

func NewHandlerRefresh(
	log *logging.Logger,
	userService *services.UserService,
) *HandlerRefresh {
	return &HandlerRefresh{
		log:         log,
		userService: userService,
	}
}

func (h *HandlerRefresh) Validate(req *Request) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerRefresh) RefreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "refresh.RefreshHandler"

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		errs := h.Validate(&req)
		if len(errs) != 0 {
			render.JSON(w, r, resp.Error(errs))
			return
		}

		t, user, err := h.userService.Refresh(req.GUID, req.RefreshToken)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		response := Response{
			Response:     resp.OK(),
			Token:        t,
			RefreshToken: user.RefreshToken,
			User:         user,
		}

		h.log.Info("User refreshed token", zap.String("op", op), zap.Any("user", user))

		render.JSON(w, r, response)
	}
}
