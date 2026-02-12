package service

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"evm_event_indexer/internal/config"
	"evm_event_indexer/internal/enum"
	"evm_event_indexer/internal/errors"
	"evm_event_indexer/internal/storage"
	"evm_event_indexer/service/model"
	adminRepo "evm_event_indexer/service/repo/admin"
	"evm_event_indexer/service/repo/session"
	userRepo "evm_event_indexer/service/repo/user"
	"evm_event_indexer/utils"
	"evm_event_indexer/utils/hashing"
)

func VerifyAdminPassword(ctx context.Context, account string, password string) (*model.Admin, error) {
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

	admins, err := adminRepo.GetAdmins(ctx, db, &adminRepo.GetAdminFilter{
		Accounts: []string{account},
		Status:   enum.UserStatusEnabled,
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get admin")
	}
	if len(admins) == 0 {
		return nil, errors.ErrInvalidCredentials.New("incorrect account or password")
	}

	admin := admins[0]
	opt := &hashing.Argon2Opt{
		Time:    admin.AuthMeta.Time,
		Memory:  admin.AuthMeta.Memory,
		Threads: admin.AuthMeta.Threads,
		KeyLen:  admin.AuthMeta.KeyLen,
	}
	if !hashing.NewArgon2(opt).Verify(password, admin.AuthMeta.Salt, admin.Password) {
		return nil, errors.ErrInvalidCredentials.New("incorrect account or password")
	}

	return admin, nil
}

func GetAdminByID(ctx context.Context, id int64) (*model.Admin, error) {
	if id <= 0 {
		return nil, errors.ErrApiInvalidParam.New("invalid admin id")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	admins, err := adminRepo.GetAdmins(ctx, db, &adminRepo.GetAdminFilter{IDs: []int64{id}})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get admin")
	}
	if len(admins) == 0 {
		return nil, errors.ErrUserNotFound.New("admin not found")
	}
	return admins[0], nil
}

func GetAdminByAccount(ctx context.Context, account string) (*model.Admin, error) {
	if account == "" {
		return nil, errors.ErrApiInvalidParam.New("account is required")
	}

	db, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	admins, err := adminRepo.GetAdmins(ctx, db, &adminRepo.GetAdminFilter{
		Accounts: []string{account},
		Status:   enum.UserStatusEnabled,
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get admin")
	}
	if len(admins) == 0 {
		return nil, nil
	}

	return admins[0], nil
}

func CreateAdminSession(ctx context.Context, adminID int64) (*model.SessionOut, error) {
	if adminID == 0 {
		return nil, errors.ErrApiInvalidParam.New("adminID")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	s, err := session.CreateAdminSession(ctx, client, adminID)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to create session")
	}

	return s, nil
}

func VerifyAdminAT(ctx context.Context, at string) (int64, error) {
	claims, err := session.VerifyAdminAccessToken(ctx, at)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err)
	}
	if claims == nil {
		return 0, errors.ErrInvalidCredentials.New("invalid access token")
	}

	adminID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to parse admin id")
	}
	return adminID, nil
}

func GetAdminIDByRT(ctx context.Context, rt string) (int64, error) {
	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	sessionID, err := session.GetAdminSessionIDByRT(ctx, client, rt)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get admin id by refresh token")
	}
	if sessionID == "" {
		return 0, errors.ErrInvalidCredentials.New("invalid refresh token")
	}

	data, err := session.GetAdminSessionData(ctx, client, sessionID)
	if err != nil {
		return 0, errors.ErrInternalServerError.Wrap(err, "failed to get session data")
	}
	if data == nil {
		return 0, errors.ErrInvalidCredentials.New("invalid refresh token")
	}

	return data.UserID, nil
}

func RevokeAdminSession(ctx context.Context, adminID int64) error {
	if adminID == 0 {
		return errors.ErrApiInvalidParam.New("admin id is 0")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	if err = session.RevokeAdminSessionByAdminID(ctx, client, adminID); err != nil {
		return errors.ErrInternalServerError.Wrap(err, "failed to revoke admin session")
	}
	return nil
}

func VerifyAdminCSRFToken(ctx context.Context, csrf string) (*model.SessionStore, error) {
	if csrf == "" {
		return nil, errors.ErrCSRFTokenInvalid.New("csrf token is required")
	}

	client, err := storage.GetRedis(config.RedisCert)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get redis")
	}

	hashed := hashing.Sha256([]byte(csrf))

	sessionID, err := session.GetAdminSessionIDByCSRF(ctx, client, hashed)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get session id by refresh token")
	}
	if sessionID == "" {
		return nil, errors.ErrInvalidCredentials.New("invalid csrf token")
	}

	data, err := session.GetAdminSessionData(ctx, client, sessionID)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get session id by refresh token")
	}
	if data == nil || data.HashedCSRF != hashed {
		return nil, errors.ErrCSRFTokenInvalid.New("invalid csrf token")
	}

	return data, nil
}

func UpdateUserByAdmin(ctx context.Context, id int64, password string, status enum.UserStatus) (*model.User, error) {
	if id <= 0 {
		return nil, errors.ErrApiInvalidParam.New("invalid user id")
	}

	dbs, err := storage.GetMySQL(config.AccountDBS)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	dbm, err := storage.GetMySQL(config.AccountDBM)
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get mysql")
	}

	users, err := userRepo.GetUsers(ctx, dbs, &userRepo.GetUserFilter{IDs: []int64{id}})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to get user")
	}

	if len(users) == 0 {
		return nil, errors.ErrUserNotFound.New()
	}

	user := users[0]

	if password != "" {
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
	}

	if status != 0 {
		user.Status = status
	}

	err = utils.NewTx(dbm).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return userRepo.TxUpdateUser(ctx, tx, user)
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to update user")
	}

	return user, nil
}

func CreateUserByAdmin(ctx context.Context, account string, password string) (*model.User, error) {
	return RegisterUser(ctx, account, password)
}

func GetUsersByAdmin(ctx context.Context, filter *userRepo.GetUserFilter) ([]*model.User, int64, error) {
	return GetUsersWithTotal(ctx, filter)
}

func DeleteUserByAdmin(ctx context.Context, id int64) error {
	return SoftDeleteUser(ctx, id)
}

func GetUserByIDByAdmin(ctx context.Context, id int64) (*model.User, error) {
	return GetUserByID(ctx, id)
}

func RegisterAdmin(ctx context.Context, account string, password string) (*model.Admin, error) {
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

	admin := &model.Admin{
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

	err = utils.NewTx(db).Exec(ctx, func(ctx context.Context, tx *sql.Tx) error {
		id, err := adminRepo.TxInsertAdmin(ctx, tx, admin)
		if err != nil {
			return err
		}
		admin.ID = id
		return nil
	})
	if err != nil {
		return nil, errors.ErrInternalServerError.Wrap(err, "failed to register admin")
	}

	return admin, nil
}
