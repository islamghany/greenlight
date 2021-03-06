package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"golang.org/x/crypto/bcrypt"
	"islamghany.greenlight/internals/marshing"
	"islamghany.greenlight/internals/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

// user model
type User struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	CreatedAt     time.Time `json:"created_at"`
	Password      password  `json:"-"`
	Activated     bool      `json:"activated"`
	Version       int       `json:"-"`
	IsCurrentUser bool      `json:"is_current_user,omitempty"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (u *User) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)

	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))

	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			{
				return false, nil
			}
		default:
			return false, err
		}
	}
	return true, nil
}

// user validation
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidateName(v *validator.Validator, name string) {
	v.Check(name != "", "name", "must be provided")
	v.Check(len(name) <= 500, "name", "must not be more than 500 bytes long")
}

func ValidatePasswordPlaintext(v *validator.Validator, passwordPlaintext string) {
	v.Check(passwordPlaintext != "", "password", "must be provided")
	v.Check(len(passwordPlaintext) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(passwordPlaintext) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	ValidateName(v, user.Name)
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

// user

type UserModel struct {
	DB  *sql.DB
	RDB *redis.Client
}

func (m UserModel) Insert(user *User) error {

	query := `
		INSERT INTO users (name, email, password_hash, activated) 
		VALUES ($1, $2, $3,$4)
		RETURNING id, created_at, version
	`

	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		From users
		WHERE email = $1
	`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil

}
func (m UserModel) GetByID(id, currentUserID int64) (*User, error) {
	query := `
		SELECT id,created_at,name, email, activated, version
		FROM users
		WHERE id = $1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var user User
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	user.IsCurrentUser = bool(currentUserID == id)
	return &user, nil
}
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	// Remember that this returns a byte *array* with length 32, not a slice.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Set up the SQL query.
	query := `
        SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	// Create a slice containing the query arguments. Notice how we use the [:] operator
	// to get a slice containing the token hash, rather than passing in the array (which
	// is not supported by the pq driver), and that we pass the current time as the
	// value to check against the token expiry.
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query, scanning the return values into a User struct. If no matching
	// record is found we return an ErrRecordNotFound error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Return the matching user.
	return &user, nil
}

func (m UserModel) CacheUserbyID(userID int64, user map[string]interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	bUser, err := marshing.MarshalBinary(user)
	err = m.RDB.Set(ctx, UsersKey(userID), bUser, time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (m UserModel) CacheRetrieveUserByID(id int64) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.RDB.Get(ctx, UsersKey(id)).Result()

	if err != nil || result == "" {
		return nil, err
	}
	return &result, nil
}
