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

// RegisterUser inserts a new user into the database (self-registration).
func RegisterUser(ctx context.Context, account string, password string) (*model.User, error) {
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
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to register user")
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

// GetUserByID retrieves a user by ID (any status).
func GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	if id <= 0 {
		return nil, errors.ErrApiInvalidParam.New("invalid user id")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	users, err := userRepo.GetUsers(ctx, db, &userRepo.GetUserFilter{IDs: []int64{id}})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}
	if len(users) == 0 {
		return nil, errors.ErrUserNotFound.New()
	}
	return users[0], nil
}

// GetUsersWithTotal returns users matching the filter and the total count.
func GetUsersWithTotal(ctx context.Context, filter *userRepo.GetUserFilter) ([]*model.User, int64, error) {
	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	total, err := userRepo.GetUserTotal(ctx, db, filter)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get user total")
	}

	if total == 0 {
		return nil, 0, nil
	}

	users, err := userRepo.GetUsers(ctx, db, filter)
	if err != nil {
		return nil, 0, errors.ErrInternalServerError.Wrap(err, "failed to get users")
	}
	return users, total, nil
}

// UpdateUser updates a user's password.
func UpdateUser(ctx context.Context, id int64, password string) (*model.User, error) {
	if id <= 0 {
		return nil, errors.ErrApiInvalidParam.New("invalid user id")
	}

	if password == "" {
		return nil, errors.ErrApiInvalidParam.New("password is required")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	dbm, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	users, err := userRepo.GetUsers(ctx, db, &userRepo.GetUserFilter{IDs: []int64{id}})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}
	if len(users) == 0 {
		return nil, errors.ErrUserNotFound.New()
	}
	user := users[0]

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

	user.Password = hashB64
	user.AuthMeta = &model.AuthMeta{
		Salt:    saltB64,
		Memory:  argonOpt.Memory,
		Time:    argonOpt.Time,
		Threads: argonOpt.Threads,
		KeyLen:  argonOpt.KeyLen,
	}

	err = utils.NewTx(dbm).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return userRepo.TxUpdateUser(ctx, tx, user)
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update user")
	}

	// Refetch to get updated_at
	users, _ = userRepo.GetUsers(ctx, db, &userRepo.GetUserFilter{IDs: []int64{id}})
	if len(users) > 0 {
		return users[0], nil
	}
	return user, nil
}

// SoftDeleteUser sets the user's status to disabled.
func SoftDeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.ErrApiInvalidParam.New("invalid user id")
	}

	db, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	// Verify user exists
	readDB, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}
	users, err := userRepo.GetUsers(ctx, readDB, &userRepo.GetUserFilter{IDs: []int64{id}})
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}
	if len(users) == 0 {
		return errors.ErrUserNotFound.New()
	}

	err = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return userRepo.TxDeleteUser(ctx, tx, id)
	})
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to delete user")
	}
	return nil
}
