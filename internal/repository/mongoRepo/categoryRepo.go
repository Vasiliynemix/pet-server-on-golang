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

var DuplicateNameError = fmt.Errorf("category with duplicate name")

type CategoryRepoM struct {
	log        *logging.Logger
	mongo      *mongodb.MongoDB
	collection string
}

func NewCategoryRepoM(log *logging.Logger, mongo *mongodb.MongoDB, collection string) *CategoryRepoM {
	return &CategoryRepoM{
		log:        log,
		mongo:      mongo,
		collection: collection,
	}
}

func (u *CategoryRepoM) CreateIndexes() error {
	const op = "CategoryRepoM.CreateIndexes"
	indexModel := mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := u.mongo.GetCollection(u.collection).Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		u.log.Error("Error creating indexes", zap.String("op", "CreateIndexes"), zap.Error(err))
		return err
	}

	u.log.Info("Indexes created", zap.String("op", op))

	return nil
}

func (u *CategoryRepoM) GetCategories() ([]*models.Category, error) {
	const op = "CategoryRepoM.GetCategories"

	collection := u.mongo.GetCollection(u.collection)

	filter := bson.D{}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		u.log.Error("Error getting categories", zap.String("op", op), zap.Error(err))
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var categories []*models.Category
	for cursor.Next(context.TODO()) {
		var category models.Category

		err = cursor.Decode(&category)
		if err != nil {
			u.log.Error("Error decoding category", zap.String("op", op), zap.Error(err))
			return nil, err
		}

		categories = append(categories, &category)
	}

	if err = cursor.Err(); err != nil {
		u.log.Error("Error getting categories", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	return categories, nil
}

func (u *CategoryRepoM) AddCategory(category *models.Category) error {
	const op = "CategoryRepoM.AddCategory"

	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.InsertOne(context.TODO(), category)
	if err != nil {
		var writeException mongo.WriteException
		if errors.As(err, &writeException) {
			return u.generateDuplicateError(writeException)
		}
		u.log.Error("Error adding category", zap.String("op", op), zap.Error(err))
		return err
	}
	return nil
}

func (u *CategoryRepoM) generateDuplicateError(err mongo.WriteException) error {
	const op = "CategoryRepoM.generateDuplicateError"

	for _, we := range err.WriteErrors {
		if we.Code == 11000 {
			u.log.Error("Category with duplicate name", zap.String("op", op), zap.Error(err))

			re := regexp.MustCompile(`"(.+?)"`)
			matches := re.FindStringSubmatch(we.Message)

			if len(matches) > 1 {
				errorMsg := fmt.Sprintf("%s: Duplicate key violation for index: %s", DuplicateNameError, matches[1])
				return errors.New(errorMsg)
			}
		}
	}

	return nil
}
