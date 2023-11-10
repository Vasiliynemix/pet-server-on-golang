package mongoRepo

import (
	"PetProjectGo/internal/models"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"time"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepoM struct {
	log        *logging.Logger
	mongo      *mongodb.MongoDB
	collection string
}

func NewUserRepoM(log *logging.Logger, mongo *mongodb.MongoDB, collection string) *UserRepoM {
	return &UserRepoM{
		log:        log,
		mongo:      mongo,
		collection: collection,
	}
}

func (u *UserRepoM) UpdateUnLoggingUser(guid string, timeNow *time.Time) error {
	const op = "UserRepoM.UpdateUnLoggingUser"

	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"guid": guid},
		bson.M{
			"$set": bson.M{
				"updated_at": timeNow,
				"is_logged":  false,
			},
		},
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrUserNotFound
		}
		u.log.Error("Error updating user refresh token", zap.String("op", op), zap.Error(err))
		return err
	}

	return nil
}

func (u *UserRepoM) UpdatedLoggingUser(guid string, rt string, timeNow *time.Time) error {
	const op = "UserRepoM.UpdatedLoggingUser"

	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.UpdateOne(
		context.TODO(),
		bson.M{"guid": guid},
		bson.M{
			"$set": bson.M{
				"refresh_token": rt,
				"last_login_at": timeNow,
				"updated_at":    timeNow,
				"is_logged":     true,
			},
		},
	)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrUserNotFound
		}
		u.log.Error("Error updating user refresh token", zap.String("op", op), zap.Error(err))
		return err
	}
	return nil
}

func (u *UserRepoM) AddUser(user *models.User) error {
	const op = "UserRepoM.AddUser"

	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		u.log.Error("Error adding user", zap.String("op", op), zap.Error(err))
		return err
	}
	return nil
}

func (u *UserRepoM) GetByLogin(login string) (*models.User, error) {
	const op = "UserRepoM.GetByLogin"
	var user *models.User

	collection := u.mongo.GetCollection(u.collection)
	err := collection.FindOne(context.TODO(), bson.M{"login": login}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		u.log.Error("Error getting user by login", zap.String("op", op), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (u *UserRepoM) GetByGuid(guid string) (*models.User, error) {
	const op = "UserRepoM.GetByGuid"
	var user *models.User

	collection := u.mongo.GetCollection(u.collection)
	err := collection.FindOne(context.TODO(), bson.M{"guid": guid}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		u.log.Error("Error getting user by guid", zap.String("op", op), zap.Error(err))
		return nil, err
	}
	return user, nil
}

func (u *UserRepoM) DeleteByGuid(guid string) error {
	const op = "UserRepoM.DeleteByGuid"
	collection := u.mongo.GetCollection(u.collection)
	_, err := collection.DeleteOne(context.TODO(), bson.M{"guid": guid})
	if err != nil {
		u.log.Error("Error deleting user by guid", zap.String("op", op), zap.Error(err))

		return err
	}

	return nil
}
