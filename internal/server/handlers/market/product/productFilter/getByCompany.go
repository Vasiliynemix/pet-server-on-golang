package productFilter

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
	CategoryId string `json:"category_id" validate:"required" mapstructure:"category_id"`
}

type ResponseProduct struct {
	resp.Response
	Products []*models.Product `json:"product"`
}

type HandlerProductGetByCompanyGuid struct {
	cfg                  *config.AppConfig
	log                  *logging.Logger
	marketProductService *services.MarketProductService
}

func NewHandlerProductGetByCompanyGuid(
	log *logging.Logger,
	marketProductService *services.MarketProductService,
) *HandlerProductGetByCompanyGuid {
	return &HandlerProductGetByCompanyGuid{
		log:                  log,
		marketProductService: marketProductService,
	}
}

func (h *HandlerProductGetByCompanyGuid) ValidateProduct(req *RequestProduct) []*handlers.ValidationError {
	return handlers.CreateValidationErrorsResp(req)
}

func (h *HandlerProductGetByCompanyGuid) AddProductGetByCompanyGuidHandler() http.HandlerFunc {
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

		products, err := h.marketProductService.GetAllByCompanyGuid(req.CategoryId)
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		render.JSON(w, r, ResponseProduct{
			Response: resp.OK(),
			Products: products,
		})
	}
}
