package user

import (
	"PetProjectGo/internal/config"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/tokenGen"
	"github.com/go-chi/render"
	"net/http"
	"strings"
)

var UnauthorizedError = "unauthorized"

type Response struct {
	resp.Response
	Token string                  `json:"new_token,omitempty"`
	User  *tokenGen.UserInfoToken `json:"user,omitempty"`
}

type HandlerUserGet struct {
	cfg         *config.AppConfig
	log         *logging.Logger
	userService *services.UserService
}

func NewHandlerUserGet(
	log *logging.Logger,
	userService *services.UserService,
) *HandlerUserGet {
	return &HandlerUserGet{
		log:         log,
		userService: userService,
	}
}

func (h *HandlerUserGet) UserGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const bearerPrefix = "Bearer "

		tokenString := r.Header.Get("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, bearerPrefix) {
			render.JSON(w, r, resp.Error(UnauthorizedError))
			return
		}

		token := strings.TrimPrefix(tokenString, bearerPrefix)
		if token == "" {
			render.JSON(w, r, resp.Error(UnauthorizedError))
			return
		}

		newT, userInfo, err := h.userService.GetMeInfo(token)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		var response Response
		if newT != "" && newT != token {
			w.Header().Set("Authorization", "Bearer "+newT)
			response = Response{
				Response: resp.OK(),
				Token:    newT,
			}
		} else {
			response = Response{
				Response: resp.OK(),
				User:     userInfo,
			}
		}

		render.JSON(w, r, response)
	}
}
