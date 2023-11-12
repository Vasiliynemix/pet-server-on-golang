package postgresRepo

import (
	"PetProjectGo/pkg/logging"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"time"
)

type UserRepoP struct {
	log      *logging.Logger
	postgres *sqlx.DB
}

func NewUserRepoP(log *logging.Logger, postgres *sqlx.DB) *UserRepoP {
	return &UserRepoP{
		log:      log,
		postgres: postgres,
	}
}

type User struct {
	GUID           string
	HashedPassword string
	CreatedAt      *time.Time
}

func (u *UserRepoP) GetHashPasswordByGuid(guid string) (string, error) {
	var hashedPassword string

	query := `SELECT hashed_password FROM passwords WHERE guid = $1`

	err := u.postgres.QueryRow(query, guid).Scan(&hashedPassword)
	return hashedPassword, err
}

func (u *UserRepoP) AddUser(nur *User) error {
	const op = "UserRepoP.AddUser"

	query := `INSERT INTO passwords (guid, hashed_password, created_at) VALUES ($1, $2, $3)`

	_, err := u.postgres.Exec(query, nur.GUID, nur.HashedPassword, nur.CreatedAt)
	if err != nil {
		u.log.Error("Error adding user", zap.String("op", op), zap.Error(err))

		return err
	}

	return nil
}
