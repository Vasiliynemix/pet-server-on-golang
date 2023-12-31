package product

import (
	"PetProjectGo/internal/config"
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

type RequestProduct struct {
	CategoryId  string `json:"category_id" validate:"required" mapstructure:"category_id"`
	Name        string `json:"name" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	Description string `json:"description" validate:"required"`
	Quantity    int    `json:"quantity,omitempty"`
}

type ResponseProduct struct {
	resp.Response
	Product *models.Product `json:"product"`
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

		var newProduct *services.NewProductM
		err = mapstructure.Decode(req, &newProduct)
		if err != nil {
			h.log.Error("Failed to parse request body", zap.String("op", op), zap.Error(err))
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		product, err := h.marketProductService.AddProduct(newProduct)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		render.JSON(w, r, ResponseProduct{
			Response: resp.OK(),
			Product:  product,
		})
	}
}
