package services

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/repository/mongoRepo"
	"PetProjectGo/pkg/storage/mongodb"
	"github.com/google/uuid"
)

type MarketCategoryService struct {
	mongo *mongoRepo.CategoryRepoM
	user  *UserService
}

func NewMarketCategoryService(
	mongo *mongodb.MongoDB,
	userService *UserService,
) (*MarketCategoryService, error) {
	mongoDb := mongoRepo.NewCategoryRepoM(userService.log, mongo, "categories")
	err := mongoDb.CreateIndexesCategory()
	if err != nil {
		return nil, err
	}
	return &MarketCategoryService{
		user:  userService,
		mongo: mongoDb,
	}, nil
}

func (c *MarketCategoryService) GetByGuid(guid string) (*models.Category, error) {
	category, err := c.mongo.GetByGuid(guid)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (c *MarketCategoryService) GetAllCategories() ([]*models.Category, error) {
	categories, err := c.mongo.GetCategories()
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (c *MarketCategoryService) AddCategory(name string) error {
	newCategory := &models.Category{
		GUID: uuid.New().String(),
		Name: name,
	}

	err := c.mongo.AddCategory(newCategory)
	if err != nil {
		return err
	}
	return nil
}
