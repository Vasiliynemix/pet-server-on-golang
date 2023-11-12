package services

import (
	"PetProjectGo/internal/repository/mongoRepo"
	"PetProjectGo/pkg/storage/mongodb"
)

type MarketProductService struct {
	mongo    *mongoRepo.ProductRepoM
	category *MarketCategoryService
}

func NewMarketProductService(
	mongo *mongodb.MongoDB,
	categoryService *MarketCategoryService,
) *MarketProductService {
	mongoDb := mongoRepo.NewProductRepoM(categoryService.user.log, mongo, "products")
	return &MarketProductService{
		category: categoryService,
		mongo:    mongoDb,
	}
}
