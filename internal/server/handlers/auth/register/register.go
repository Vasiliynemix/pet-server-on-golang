package register

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/server/handlers"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"fmt"
	"github.com/go-chi/render"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"net/http"
)

type Request struct {
	Login             string `json:"login" validate:"required"`
	Password          string `json:"password" validate:"required"`
	ConfirmedPassword string `json:"confirmed_password" validate:"required,eqfield=Password"`
	Name              string `json:"name,omitempty"`
	LastName          string `json:"last_name,omitempty"`
}

type Response struct {
	resp.Response
	User *models.User `json:"user"`
}

type HandlerRegister struct {
	cfg         *config.AppConfig
	log         *logging.Logger
	userService *services.UserService
}

func NewHandlerRegister(
	cfg *config.AppConfig,
	log *logging.Logger,
	userService *services.UserService,
) *HandlerRegister {
	return &HandlerRegister{
		cfg:         cfg,
		log:         log,
		userService: userService,
	}
}

func (h *HandlerRegister) Validate(req *Request, passwordMinLength int) []*handlers.ValidationError {
	errs := handlers.CreateValidationErrorsResp(req)

	if len(req.Password) < passwordMinLength {
		var element handlers.ValidationError
		element.Field = "Request.Password"
		element.Tag = fmt.Sprintf("min length password: %d", passwordMinLength)
		element.Value = ""
		errs = append(errs, &element)
	}

	return errs
}

func (h *HandlerRegister) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "register.RegisterHandler"

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		errs := h.Validate(&req, h.cfg.PasswordMinLength)
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

		user, err := h.userService.Register(newUser)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		response := Response{
			Response: resp.OK(),
			User:     user,
		}

		h.log.Info("User registered", zap.String("op", op), zap.Any("user", user))

		render.JSON(w, r, response)
	}
}
