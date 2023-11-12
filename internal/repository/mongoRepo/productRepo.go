package mongoRepo

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
	"fmt"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"regexp"
)

var DuplicateProductNameError = fmt.Errorf("product with duplicate name")

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

func (u *ProductRepoM) CreateIndexesProduct() error {
	const op = "ProductRepoM.CreateIndexesProduct"

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := u.mongo.GetCollection(u.collection).Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		u.log.Error("Error creating indexes", zap.String("op", op), zap.Error(err))
		return err
	}

	u.log.Debug("Indexes product created", zap.String("op", op))

	return nil
}

func (u *ProductRepoM) AddNewProduct(product *models.Product) error {
	const op = "ProductRepoM.AddProduct"
	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.InsertOne(context.TODO(), product)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			return u.generateDuplicateErrorP(writeException)
		}
		u.log.Error("Error adding product", zap.String("op", op), zap.Error(err))
		return err
	}
	return nil
}

func (u *ProductRepoM) GetProductsByCategoryGuid(categoryGuid string) ([]*models.Product, error) {
	const op = "ProductRepoM.GetProductsByCategoryGuid"
	collection := u.mongo.GetCollection(u.collection)

	filter := bson.M{
		"category_id": categoryGuid,
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		u.log.Error("Error getting products", zap.String("op", op), zap.Error(err))
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var products []*models.Product
	for cursor.Next(context.TODO()) {
		var product models.Product

		err = cursor.Decode(&product)
		if err != nil {
			u.log.Error("Error decoding product", zap.String("op", op), zap.Error(err))
			return nil, err
		}

		products = append(products, &product)
	}

	if err = cursor.Err(); err != nil {
		u.log.Error("Error getting products", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return products, nil
}

func (u *ProductRepoM) generateDuplicateErrorP(err mongo.WriteException) error {
	const op = "ProductRepoM.generateDuplicateErrorP"

	for _, we := range err.WriteErrors {
		if we.Code == 11000 {
			u.log.Error("Product with duplicate name", zap.String("op", op), zap.Error(err))

			re := regexp.MustCompile(`"(.+?)"`)
			matches := re.FindStringSubmatch(we.Message)

			if len(matches) > 1 {
				errorMsg := fmt.Sprintf("%s: Duplicate key violation for index: %s", DuplicateProductNameError, matches[1])
				return errors.New(errorMsg)
			}
		}
	}

	return nil
}
