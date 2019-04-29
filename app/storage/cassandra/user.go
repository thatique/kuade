package cassandra

import (
	"github.com/gocql/gocql"

	"github.com/thatique/app/store/cassandra/dbmodels"
	"github.com/thatique/kuade/app/model"
	"github.com/thatique/kuade/pkg/kerr"
)

const (
	insertUser = `INSERT INTO users(id, slug, name, email, icon, role, status, bio, age, address, category, budget, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertUserSlug = `INSERT INTO users_by_slug(id, slug, name, email, icon, role, status, bio, age, address, category, budget, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	insertUserCredential = `INSERT INTO user_credentials(email, user_id, password, enabled, created_at, last_signin)
		VALUES (?, ?, ?, ?, ?, ?)`
	insertUserProvider = `INSERT INTO user_providers(name, key, user_id) VALUES (?, ?, ?)`

	queryUserByProvider        = `SELECT user_id FROM user_providers WHERE name = ? AND key = ?`
	queryUserCredentialByEmail = `SELECT user_id, password, enabled, created_at, last_signin FROM user_credentials WHERE email = ?`
	queryUserByID              = `SELECT id, slug, name, email, icon, role, status, bio, age, address, category, budget, created_at
		FROM users
		WHERE id = ?`
	queryUserBySlug = `SELECT id, slug, name, email, icon, role, status, bio, age, address, category, budget, created_at
		FROM users_by_slug
		WHERE slug = ?`
)

type userStore struct {
	session *gocql.Session
}

func (s *userStore) PutUser(ctx context.Context, user *model.User) error {
	usr := dbmodels.FromDomainUser(user)
	if err := s.putUser(ctx, usr); err != nil {
		return err
	}
	if err := s.putUserSlug(ctx, usr); err != nil {
		return err
	}
	return nil
}

func (s *userStore) PutUserCredential(ctx context.Context, id model.ID, creds *model.Credentials) error {
	dbCreds := dbmodels.FromDomainUserCredential(id, creds)
	query := s.session.Query(insertUserCreds, dbCreds.Email, dbCreds.UserID, dbCreds.Passwords,
		dbCreds.Enabled, dbCreds.CreatedAt, dbCreds.LastSignin).WithContext(ctx)
	return query.Exec()
}

func (s *userStore) FindOrCreateUserForProvider(ctx context.Context, userData *model.User, provider model.OauthProvider) (newUser bool, user *model.User, err error) {
	var userId int64
	if err := s.session.Query(queryUserByProvider, provider.Name, provider.Key).WithContext(ctx).Scan(&userId); err != nil {
		if err != gocql.ErrNotFound {
			return false, nil, err
		}
		// insert it
		if err = s.PutUser(ctx, userData); err != nil {
			return false, nil, err
		}

		if err = s.session.Query(insertUserProvider, provider.Name, provider.Key, int64(userData.ID)).WithContext(ctx).Exec(); err != nil {
			return false, nil, err
		}
		return true, userData, nil
	}
	return s.getUserById(ctx, userId)
}

func (s *userStore) IsEmailAlreadyInUse(ctx context.Context, email string) (bool, model.ID, error) {
	creds, err := s.getUserCredential(ctx, email)
	if err != nil {
		if err == gocql.ErrNotFound {
			return true, model.ID(creds.UserID), nil
		}
		return false, model.ID(0), err
	}
	return true, model.ID(creds.UserID), nil
}

func (s *userStorage) GetCredentialByEmail(ctx context.Context, email string) (*model.Credentials, error) {
	creds, err := s.getUserCredential(ctx, email)
	if err != nil {
		return nil, err
	}
	return creds.ToDomain(), nil
}

func (s *userStore) GetUserByID(ctx context.Context, id model.ID) (*model.User, error) {
	user, err := s.getUserById(ctx, id)
	if err != nil {
		return nil, err
	}
	return user.ToDomain(), nil
}

func (s *userStore) GetUserBySlug(ctx context.Context, slug string) (*model.User, error) {
	usr = &dbmodels.User{}
	if err := s.session.Query(queryUserBySlug, slug).WithContext(ctx).Scan(
		&usr.ID,
		&usr.Slug,
		&usr.Name,
		&usr.Email,
		&usr.Icon,
		&usr.Role,
		&usr.Status,
		&usr.Bio,
		&usr.Age,
		&usr.Address,
		&usr.Category,
		&usr.Budget,
		&usr.CreatedAt,
	); err != nil {
		return nil, err
	}
	return usr.ToDomain(), nil
}

func (s *userStorage) ErrorCode(err error) kerr.ErrorCode {
	if err == gocql.ErrNotFound {
		return kerr.NotFound
	}

	return kerr.Unknown
}

func (s *userStore) getUserCredential(ctx context.Context, email string) (*dbmodels.Credentials, error) {
	creds := &dbmodels.Credentials{Email: email}
	if err := s.session.Query(queryUserCredentialByEmail, email).WithContext(ctx).Scan(
		&creds.UserID,
		&creds.Passwords,
		&creds.Enabled,
		&creds.CreatedAt,
		&creds.LastSignin,
	); err != nil {
		return nil, err
	}
	return creds, nil
}

func (s *userStore) putUser(ctx context.Context, usr *dbmodels.User) error {
	query := s.session.Query(insertUser, usr.ID, usr.Slug, usr.Name, usr.Email, usr.Icon,
		usr.Role, usr.Status, usr.Bio, usr.Age, usr.Address, usr.Category, usr.Budget, usr.CreatedAt).WithContext(ctx)
	return query.Exec()
}

func (s *userStore) putUserSlug(ctx context.Context, usr *dbmodels.User) error {
	query := s.session.Query(insertUserSlug, usr.ID, usr.Slug, usr.Name, usr.Email, usr.Icon,
		usr.Role, usr.Status, usr.Bio, usr.Age, usr.Address, usr.Category, usr.Budget, usr.CreatedAt).WithContext(ctx)
	return query.Exec()
}

func (s *userStore) getUserById(ctx context.Context, id int64) (user *dbmodels.User, err error) {
	usr = &dbmodels.User{}
	if err := s.session.Query(queryUserByID, userId).WithContext(ctx).Scan(
		&usr.ID,
		&usr.Slug,
		&usr.Name,
		&usr.Email,
		&usr.Icon,
		&usr.Role,
		&usr.Status,
		&usr.Bio,
		&usr.Age,
		&usr.Address,
		&usr.Category,
		&usr.Budget,
		&usr.CreatedAt,
	); err != nil {
		return nil, err
	}
	return usr, nil
}
