package unlogin

import (
	"PetProjectGo/internal/server/handlers"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"github.com/go-chi/render"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"net/http"
)

type Request struct {
	GUID string `json:"id" validate:"required"`
}

type Response struct {
	resp.Response
}

type HandlerUnLogin struct {
	log         *logging.Logger
	userService *services.UserService
}

func NewHandlerUnLogin(
	log *logging.Logger,
	userService *services.UserService,
) *HandlerUnLogin {
	return &HandlerUnLogin{
		log:         log,
		userService: userService,
	}
}

func (h *HandlerUnLogin) Validate(req *Request) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerUnLogin) UnLoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "login.LoginHandler"

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

		var newUser *services.NewUserM
		err = mapstructure.Decode(req, &newUser)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		err = h.userService.UnLogin(req.GUID)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		response := Response{
			Response: resp.OK(),
		}

		h.log.Info("User unlogged", zap.String("op", op), zap.String("guid", req.GUID))

		render.JSON(w, r, response)
	}
}
