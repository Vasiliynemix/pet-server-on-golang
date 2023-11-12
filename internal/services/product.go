package services

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/repository/mongoRepo"
	"PetProjectGo/pkg/storage/mongodb"
	"github.com/google/uuid"
)

type NewProductM struct {
	CategoryGuid string `json:"category_id" mapstructure:"category_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int    `json:"price"`
	Quantity     int    `json:"quantity,omitempty" default:"1"`
}

type MarketProductService struct {
	mongo    *mongoRepo.ProductRepoM
	category *MarketCategoryService
}

func NewMarketProductService(
	mongo *mongodb.MongoDB,
	categoryService *MarketCategoryService,
) (*MarketProductService, error) {
	mongoDb := mongoRepo.NewProductRepoM(categoryService.user.log, mongo, "products")
	err := mongoDb.CreateIndexesProduct()
	if err != nil {
		return nil, err
	}
	return &MarketProductService{
		category: categoryService,
		mongo:    mongoDb,
	}, nil
}

func (p *MarketProductService) GetAllByCompanyGuid(guid string) ([]*models.Product, error) {
	_, err := p.category.GetByGuid(guid)
	if err != nil {
		return nil, err
	}

	products, err := p.mongo.GetProductsByCategoryGuid(guid)
	if err != nil {
		return nil, err
	}
	if products == nil {
		return []*models.Product{}, nil
	}

	return products, nil
}

func (p *MarketProductService) AddProduct(product *NewProductM) (*models.Product, error) {
	_, err := p.category.GetByGuid(product.CategoryGuid)
	if err != nil {
		return nil, err
	}
	userGuid := uuid.New().String()
	newProduct := &models.Product{
		GUID:         userGuid,
		CategoryGuid: product.CategoryGuid,
		Name:         product.Name,
		Description:  product.Description,
		Price:        product.Price,
		Quantity:     product.Quantity,
	}
	err = p.mongo.AddNewProduct(newProduct)
	if err != nil {
		return nil, err
	}
	return newProduct, nil
}
