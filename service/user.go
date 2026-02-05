package service

import (
	"context"
	"database/sql"
	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	userRepo "evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"
	"time"
)

// VerifyUserPassword verifies the user password, return InvalidCredentialsError if the account or password is incorrect.
func VerifyUserPassword(ctx context.Context, account string, password string) (*model.User, error) {

	if account == "" {
		return nil, errors.ErrApiInvalidParam.New("account is required")
	}
	if password == "" {
		return nil, errors.ErrApiInvalidParam.New("password is required")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	users, err := userRepo.GetUsers(ctx, db, &userRepo.GetUserFilter{
		Accounts: []string{account},
		Status:   enum.UserStatusEnabled,
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}

	if len(users) == 0 {
		return nil, errors.ErrInvalidCredentials.New("incorrect account or password")
	}

	user := users[0]

	opt := &hashing.Argon2Opt{
		Time:    user.AuthMeta.Time,
		Memory:  user.AuthMeta.Memory,
		Threads: user.AuthMeta.Threads,
		KeyLen:  user.AuthMeta.KeyLen,
	}

	if !hashing.NewArgon2(opt).Verify(password, user.AuthMeta.Salt, user.Password) {
		return nil, errors.ErrInvalidCredentials.New("incorrect account or password")
	}

	return user, nil
}

// InsertUser inserts a new user into the database
func InsertUser(ctx context.Context, account string, password string) (*model.User, error) {
	if account == "" {
		return nil, errors.ErrApiInvalidParam.New("account is required")
	}
	if password == "" {
		return nil, errors.ErrApiInvalidParam.New("password is required")
	}

	now := time.Now()

	argonOpt := &hashing.Argon2Opt{
		Time:    config.Get().Argon2.Time,
		Memory:  config.Get().Argon2.Memory,
		Threads: config.Get().Argon2.Threads,
		KeyLen:  config.Get().Argon2.KeyLen,
	}

	hasher := hashing.NewArgon2(argonOpt)
	hashB64, saltB64, err := hasher.Hashing(password)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to hash password")
	}

	user := &model.User{
		Account:  account,
		Status:   enum.UserStatusEnabled,
		Password: hashB64,
		AuthMeta: &model.AuthMeta{
			Salt:    saltB64,
			Memory:  argonOpt.Memory,
			Time:    argonOpt.Time,
			Threads: argonOpt.Threads,
			KeyLen:  argonOpt.KeyLen,
		},
		CreatedAt: now,
	}
	db, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	err = utils.NewTx(db).Exec(ctx,
		func(ctx context.Context, tx *sql.Tx) error {
			id, err := userRepo.TxInsertUser(ctx, tx, user)
			if err != nil {
				return err
			}
			user.ID = id
			return nil
		})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to insert user")
	}

	return user, nil
}

// GetUserByAccount retrieves a user by account
func GetUserByAccount(ctx context.Context, account string) (*model.User, error) {
	if account == "" {
		return nil, errors.ErrApiInvalidParam.New("account is required")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	users, err := userRepo.GetUsers(ctx, db, &userRepo.GetUserFilter{
		Accounts: []string{account},
		Status:   enum.UserStatusEnabled,
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}
