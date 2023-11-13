package handlers

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/models"
	resp "PetProjectGo/internal/server/handlers/response"
	"PetProjectGo/internal/services"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/tokenGen"
	"github.com/go-chi/render"
	"go.uber.org/zap"
	"html/template"
	"net/http"
)

type Response struct {
	resp.Response
	Token string                  `json:"new_token,omitempty"`
	User  *tokenGen.UserInfoToken `json:"user,omitempty"`
}

type CategoryData struct {
	Category *models.Category
	Products []*models.Product
}

type PageData struct {
	Categories map[string]CategoryData
	Users      []*models.User
}

type HandlerIndex struct {
	cfg             *config.AppConfig
	log             *logging.Logger
	userService     *services.UserService
	categoryService *services.MarketCategoryService
	productService  *services.MarketProductService
}

func NewHandlerIndex(
	log *logging.Logger,
	userService *services.UserService,
	categoryService *services.MarketCategoryService,
	productService *services.MarketProductService,
) *HandlerIndex {
	return &HandlerIndex{
		log:             log,
		userService:     userService,
		categoryService: categoryService,
		productService:  productService,
	}
}

func (h *HandlerIndex) IndexHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		categoryData := make(map[string]CategoryData)

		categories, err := h.categoryService.GetAllCategories()
		if err != nil {
			render.JSON(w, r, resp.Error(err.Error()))
			return
		}

		var products []*models.Product
		for _, category := range categories {
			products, err = h.productService.GetAllByCompanyGuid(category.GUID)
			if err != nil {
				render.JSON(w, r, resp.Error(err.Error()))
				return
			}
			categoryData[category.GUID] = CategoryData{Category: category, Products: products}
		}

		users, err := h.userService.GetAllUsers()
		pageData := PageData{
			Categories: categoryData,
			Users:      []*models.User{},
		}
		pageData.Users = users

		h.log.Info("data", zap.Any("data", pageData))

		files := []string{
			"./templates/index.html",
			"./templates/base.html",
			"./templates/footer.html",
			"./templates/header.html",
			"./templates/registrationForm.html",
			"./templates/categories.html",
			"./templates/products.html",
			"./templates/users.html",
		}

		ts, err := template.ParseFiles(files...)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = ts.Execute(w, pageData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}
