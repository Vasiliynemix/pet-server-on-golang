package login

import (
	"PetProjectGo/internal/models"
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
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	Token        string       `json:"token"`
	RefreshToken string       `json:"refresh_token"`
	User         *models.User `json:"user"`
}

type HandlerLogin struct {
	log         *logging.Logger
	userService *services.UserService
}

func NewHandlerLogin(
	log *logging.Logger,
	userService *services.UserService,
) *HandlerLogin {
	return &HandlerLogin{
		log:         log,
		userService: userService,
	}
}

func (h *HandlerLogin) Validate(req *Request) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerLogin) LoginHandler() http.HandlerFunc {
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

		t, user, err := h.userService.Login(req.Login, req.Password)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		w.Header().Set("Authorization", "Bearer "+t)

		response := Response{
			Response:     resp.OK(),
			Token:        t,
			RefreshToken: user.RefreshToken,
			User:         user,
		}

		h.log.Info("User logged", zap.String("op", op), zap.Any("user", user))

		render.JSON(w, r, response)
	}
}
