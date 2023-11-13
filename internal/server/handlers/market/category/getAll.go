package category

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/models"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"github.com/go-chi/render"
	"net/http"
)

type ResponseCategoryAll struct {
	resp.Response
	Categories []*models.Category `json:"categories"`
}

type HandlerCategoryAll struct {
	cfg                   *config.AppConfig
	log                   *logging.Logger
	marketCategoryService *services.MarketCategoryService
}

func NewHandlerCategoryAll(
	log *logging.Logger,
	marketCategoryService *services.MarketCategoryService,
) *HandlerCategoryAll {
	return &HandlerCategoryAll{
		log:                   log,
		marketCategoryService: marketCategoryService,
	}
}

func (h *HandlerCategoryAll) AllCategoriesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := h.marketCategoryService.GetAllCategories()
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		render.JSON(w, r, ResponseCategoryAll{
			Response:   resp.OK(),
			Categories: categories,
		})

		render.HTML(w, r, "categories")
	}
}
