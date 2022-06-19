package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	"islamghany.greenlight/internals/marshing"
	"islamghany.greenlight/internals/validator"
)

// JSON-encoded output.
type Movie struct {
	ID               int64     `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Title            string    `json:"title"`
	Year             int32     `json:"year,omitempty"`
	Runtime          Runtime   `json:"runtime,omitempty"`
	Genres           []string  `json:"genres"`
	Version          int32     `json:"version"`
	Count            int32     `json:"count,omitempty"`
	Likes            int64     `json:"likes,omitempty"`
	UserName         string    `json:"username"`
	UserID           int64     `json:"user_id"`
	CurrentUserLiked int32     `json:"currentUserLiked,omitempty,"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB  *sql.DB
	RDB *redis.Client
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(), movie_id as id, movies.created_at  as created_at, title, year, runtime, genres, movies.version as version, views.count as view_count, users.name as username, user_id,
		(SELECT count(*) FROM likes WHERE movies.id = likes.movie_id) as likes
	FROM movies
	INNER JOIN views ON id = views.movie_id
	INNER JOIN users ON user_id = users.id
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	AND (genres @> $2 OR $2 = '{}')     
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	movies := []*Movie{}

	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
			&movie.Count,
			&movie.UserName,
			&movie.UserID,
			&movie.Likes,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return movies, metadata, nil
}
func (m MovieModel) Insert(movie *Movie, userID int64) error {

	query := `
		INSERT INTO movies (title, year, runtime, genres,user_id)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, version
	`

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), userID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
	if err != nil {
		return err
	}
	query = `
		INSERT INTO views (movie_id)
		VALUES ($1);
	`
	m.DB.QueryRowContext(ctx, query, movie.ID)

	return nil
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT
  movies.id as id,
  movies.created_at as created_at,
  title,
  year,
  runtime,
  genres,
  movies.version as version,
  views.count as count,
  coalesce(A.count,0) as likesCount,
  users.name as username,
  movies.user_id,
  coalesce(A.currentUserLiked,0) as currentUserLiked
FROM
  movies
  inner join views on views.movie_id = movies.id
  inner join users on users.id = user_id
  LEFT JOIN (
    SELECT
      COUNT(*) as count,
      A.movie_id,
      sum(
        CASE
          WHEN A.user_id = user_id then 1
          else 0
        end
      ) as currentUserLiked
    FROM
      likes A
    GROUP BY
      A.movie_id
  ) A ON A.movie_id = movies.id
WHERE
  movies.id = $1;
	`

	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
		&movie.Count,
		&movie.Likes,
		&movie.UserName,
		&movie.UserID,
		&movie.CurrentUserLiked,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	query = `
	UPDATE views
	SET count = count+1
	WHERE movie_id = $1;
	`
	m.DB.QueryRowContext(ctx, query, id)
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {

	query := `
		UPDATE movies
		SET title = $1,  year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version=$6
		RETURNING version
	`
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := m.DB.ExecContext(ctx, query, id)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m UserModel) Update(user *User) error {
	query := `
        UPDATE users 
        SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
        WHERE id = $5 AND version = $6
        RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) GetMostViews() ([]*Movie, error) {
	query := `
	select m.id as id, created_at, title, year, runtime, genres, version, user_id, count from movies m 
	join "views" v on v.movie_id = m.id
	order by v.count desc
	limit 10;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	movies := []*Movie{}

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var movie Movie

		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
			&movie.UserID,
			&movie.Count,
		)
		if err != nil {
			return nil, err
		}

		// Add the Movie struct to the slice.
		movies = append(movies, &movie)

	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return movies, nil
}

func (m *MovieModel) CacheMost20PercentageView() error {
	fmt.Println("begin")
	query := `
		SELECT count(*) as number from movies m
		join "views" v on v.movie_id = m.id;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var totalMovies int64
	err := m.DB.QueryRowContext(ctx, query).Scan(&totalMovies)
	if err != nil {
		return err
	}

	limit := 10
	//math.Round(float64(20 / 100 * totalMovies))

	query = `
	select  m.id as id, created_at, title, year, runtime, genres, version, user_id, count from movies m 
	join "views" v on v.movie_id = m.id
	order by v.count desc
	limit $1
	`

	var jsonMovies [][]byte

	rows, err := m.DB.QueryContext(ctx, query, limit)

	if err != nil {
		return err
	}
	defer rows.Close()
	fmt.Println("fetch the movies")
	var keys []string
	for rows.Next() {
		var movie Movie
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
			&movie.UserID,
			&movie.Count,
		)
		if err != nil {
			return err
		}
		keys = append(keys, MoviesKey(movie.ID))
		jsonMovie, err := marshing.MarshalBinary(movie)
		if err != nil {
			return err
		}
		jsonMovies = append(jsonMovies, jsonMovie)

	}
	fmt.Println("full arrays")
	if err = rows.Err(); err != nil {
		return err
	}

	err = PipeSet(m.RDB, keys, jsonMovies, 3*time.Hour)
	fmt.Println("finish successfully")
	return err

}
