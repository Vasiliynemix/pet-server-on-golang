package mongoRepo

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
)

type ProductRepoM struct {
	log        *logging.Logger
	mongo      *mongodb.MongoDB
	collection string
}

func NewProductRepoM(log *logging.Logger, mongo *mongodb.MongoDB, collection string) *ProductRepoM {
	return &ProductRepoM{
		log:        log,
		mongo:      mongo,
		collection: collection,
	}
}

func (u *ProductRepoM) AddProduct(categoryId string, product *models.Product) error {
	const op = "ProductRepoM.AddProduct"

	return nil
}
