package services

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/models"
	"PetProjectGo/internal/repository/mongoRepo"
	"PetProjectGo/internal/repository/postgresRepo"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
	"PetProjectGo/pkg/tokenGen"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

const userCollection = "users"

var ErrUserAlreadyExists = fmt.Errorf("user already exists")
var InvalidLoginPassword = fmt.Errorf("invalid login or password")
var RefreshTokenExpiredError = fmt.Errorf("refresh token expired")
var UserIsLogged = fmt.Errorf("user is logged")
var UserIsUnLogged = fmt.Errorf("user is unlogged")
var Unauthorized = fmt.Errorf("unauthorized")

type NewUserM struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Name     string `json:"name"`
	LastName string `json:"last_name"`
}

type UserService struct {
	log      *logging.Logger
	cfg      *config.AppConfig
	mongo    *mongoRepo.UserRepoM
	postgres *postgresRepo.UserRepoP
}

func NewUserService(
	log *logging.Logger,
	cfg *config.AppConfig,
	mongo *mongodb.MongoDB,
	postgres *sqlx.DB,
) *UserService {
	mongoDb := mongoRepo.NewUserRepoM(log, mongo, userCollection)
	postgresDb := postgresRepo.NewUserRepoP(log, postgres)
	return &UserService{
		log:      log,
		cfg:      cfg,
		mongo:    mongoDb,
		postgres: postgresDb,
	}
}

func (u *UserService) GetAllUsers() ([]*models.User, error) {
	users, err := u.mongo.GetAllUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserService) GetMeInfo(token string) (string, *tokenGen.UserInfoToken, error) {
	userInfoToken, ok := tokenGen.VerifyToken(u.cfg.SecretKeyToken, token)
	if userInfoToken == nil {
		return "", nil, Unauthorized
	}
	user, _ := u.mongo.GetByGuid(userInfoToken.ID)
	if !user.IsLogged {
		return "", nil, UserIsUnLogged
	}

	if !ok {
		return u.checkTokenExpired(user)
	}

	return token, userInfoToken, nil
}

func (u *UserService) UnLogin(guid string) error {
	user, err := u.mongo.GetByGuid(guid)
	if errors.Is(err, mongoRepo.ErrUserNotFound) {
		return InvalidLoginPassword
	}
	if err != nil {
		return err
	}

	if !user.IsLogged {
		return UserIsUnLogged
	}

	timeNow := time.Now()
	err = u.mongo.UpdateUnLoggingUser(guid, &timeNow)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserService) Login(login string, password string) (string, *models.User, error) {
	user, err := u.mongo.GetByLogin(login)
	if errors.Is(err, mongoRepo.ErrUserNotFound) {
		return "", nil, InvalidLoginPassword
	}
	if err != nil {
		return "", nil, err
	}

	_, ok := tokenGen.VerifyToken(u.cfg.SecretKeyToken, user.RefreshToken)
	if ok {
		if user.IsLogged {
			return "", nil, UserIsLogged
		}
	}

	hashedPassword, err := u.postgres.GetHashPasswordByGuid(user.GUID)
	if err != nil {
		return "", nil, InvalidLoginPassword
	}

	ok = u.checkHashPassword(password, hashedPassword)
	if !ok {
		return "", nil, InvalidLoginPassword
	}

	t, rt, err := u.generateTokens(user)

	timeNow := time.Now()
	err = u.mongo.UpdatedLoggingUser(user.GUID, rt, &timeNow)
	if err != nil {
		return "", nil, err
	}

	user.LastLoginAt = &timeNow
	user.RefreshToken = rt
	user.IsLogged = true

	return t, user, nil
}

func (u *UserService) Refresh(guid string) (string, *models.User, error) {
	user, err := u.mongo.GetByGuid(guid)
	if user == nil {
		return "", nil, mongoRepo.ErrUserNotFound
	}
	if err != nil {
		if !errors.Is(err, mongoRepo.ErrUserNotFound) {
			return "", nil, err
		}
	}

	_, ok := tokenGen.VerifyToken(u.cfg.SecretKeyToken, user.RefreshToken)
	if !ok {
		return "", nil, RefreshTokenExpiredError
	}

	var newInfoToken *tokenGen.UserInfoToken
	err = mapstructure.Decode(user, &newInfoToken)
	if err != nil {
		return "", nil, err
	}

	timeTExpired := time.Now().Add(u.cfg.TokenExpirationTimeMinutes * time.Minute)
	t, err := tokenGen.NewToken(u.cfg.SecretKeyToken, timeTExpired, newInfoToken)

	timeNow := time.Now()
	err = u.mongo.UpdatedLoggingUser(user.GUID, user.RefreshToken, &timeNow)
	if err != nil {
		return "", nil, err
	}

	user.LastLoginAt = &timeNow

	return t, user, nil
}

func (u *UserService) Register(nur *NewUserM) (*models.User, error) {
	const op = "UserService.Register"

	user, err := u.mongo.GetByLogin(nur.Login)
	if user != nil {
		u.log.Info("User already exists", zap.String("op", op), zap.Error(ErrUserAlreadyExists))
		return nil, ErrUserAlreadyExists
	}
	if err != nil {
		if !errors.Is(err, mongoRepo.ErrUserNotFound) {
			return nil, err
		}
	}

	timeNow := time.Now()
	userGuid := uuid.New().String()
	newUser := &models.User{
		GUID:      userGuid,
		Login:     nur.Login,
		Name:      nur.Name,
		LastName:  nur.LastName,
		CreatedAt: &timeNow,
	}

	err = u.mongo.AddUser(newUser)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := u.HashPassword(nur.Password)
	if err != nil {
		return nil, err
	}
	newUserP := postgresRepo.User{
		GUID:           userGuid,
		HashedPassword: hashedPassword,
		CreatedAt:      &timeNow,
	}

	err = u.postgres.AddUser(&newUserP)
	if err != nil {
		err2 := u.mongo.DeleteByGuid(userGuid)
		if err2 != nil {
			u.log.Error("Error deleting user", zap.String("op", op), zap.Error(err))
			return nil, err2
		}
		return nil, err
	}
	return newUser, err
}

func (u *UserService) checkHashPassword(password string, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (u *UserService) HashPassword(password string) (string, error) {
	const op = "UserService.HashPassword"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		u.log.Error("Error hashing password", zap.String("op", op), zap.Error(err))
		return "", err
	}
	return string(hashed), nil
}

func (u *UserService) generateTokens(user *models.User) (string, string, error) {
	var newInfoToken *tokenGen.UserInfoToken

	timeTExpired := time.Now().Add(u.cfg.TokenExpirationTimeMinutes * time.Minute)
	timeRtExpired := time.Now().Add(u.cfg.RefreshTokenExpirationTimeMinutes * time.Minute)

	err := mapstructure.Decode(user, &newInfoToken)
	if err != nil {
		return "", "", err
	}

	t, err := tokenGen.NewToken(u.cfg.SecretKeyToken, timeTExpired, newInfoToken)
	if err != nil {
		return "", "", err
	}

	rt, err := tokenGen.NewToken(u.cfg.SecretKeyToken, timeRtExpired, nil)
	if err != nil {
		return "", "", err
	}

	return t, rt, nil
}

func (u *UserService) checkTokenExpired(user *models.User) (string, *tokenGen.UserInfoToken, error) {
	const op = "UserService.checkTokenExpired"

	var NewUserInfoToken *tokenGen.UserInfoToken
	err := mapstructure.Decode(user, &NewUserInfoToken)
	if err != nil {
		u.log.Error("can't decode user", zap.String("op", op), zap.Error(err))
		return "", nil, err
	}

	t, _, err := u.Refresh(user.GUID)
	if err != nil {
		if errors.Is(err, RefreshTokenExpiredError) {
			hashedPassword, errGetPassword := u.postgres.GetHashPasswordByGuid(user.GUID)
			if errGetPassword != nil {
				return "", nil, errGetPassword
			}
			newT, _, errReLogin := u.Login(user.Login, hashedPassword)
			if errReLogin != nil {
				return "", nil, errReLogin
			}
			return newT, NewUserInfoToken, nil
		}
		return "", nil, err
	}

	return t, NewUserInfoToken, nil
}
