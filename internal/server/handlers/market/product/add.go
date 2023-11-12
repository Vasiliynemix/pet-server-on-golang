package product

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

type RequestProduct struct {
	Name string `json:"name" validate:"required"`
}

type ResponseProduct struct {
	resp.Response
	Categories []*models.Category `json:"categories"`
}

type HandlerProductAdd struct {
	cfg                  *config.AppConfig
	log                  *logging.Logger
	marketProductService *services.MarketProductService
}

func NewHandlerProductAdd(
	log *logging.Logger,
	marketProductService *services.MarketProductService,
) *HandlerProductAdd {
	return &HandlerProductAdd{
		log:                  log,
		marketProductService: marketProductService,
	}
}

func (h *HandlerProductAdd) ValidateProduct(req *RequestProduct) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerProductAdd) AddProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "category.AddCategoryHandler"

		var req RequestProduct

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		errs := h.ValidateProduct(&req)
		if len(errs) != 0 {
			render.JSON(w, r, resp.Error(errs))
			return
		}
		err = h.marketProductService.AddProduct(req.Name)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		categories, err := h.marketCategoryService.GetAllCategories()
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}
		render.JSON(w, r, ResponseProduct{
			Response:   resp.OK(),
			Categories: categories,
		})
	}
}
