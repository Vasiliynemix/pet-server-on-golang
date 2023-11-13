package category

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/server/handlers"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"net/http"
)

type RequestCategory struct {
	Name string `json:"name" validate:"required"`
}

type ResponseCategory struct {
	resp.Response
	Categories []*models.Category `json:"categories"`
}

type HandlerCategoryAdd struct {
	cfg                   *config.AppConfig
	log                   *logging.Logger
	marketCategoryService *services.MarketCategoryService
}

func NewHandlerCategoryAdd(
	log *logging.Logger,
	marketCategoryService *services.MarketCategoryService,
) *HandlerCategoryAdd {
	return &HandlerCategoryAdd{
		log:                   log,
		marketCategoryService: marketCategoryService,
	}
}

func (h *HandlerCategoryAdd) ValidateCategory(req *RequestCategory) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerCategoryAdd) AddCategoryHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "category.AddCategoryHandler"

		var req RequestCategory

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		errs := h.ValidateCategory(&req)
		if len(errs) != 0 {
			render.JSON(w, r, resp.Error(errs))
			return
		}
		err = h.marketCategoryService.AddCategory(req.Name)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		categories, err := h.marketCategoryService.GetAllCategories()
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		render.JSON(w, r, ResponseCategory{
			Response:   resp.OK(),
			Categories: categories,
		})
	}
}
