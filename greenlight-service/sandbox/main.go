package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
	"islamghany.greenlight/internals/mailer"
)

var (
	ctx = context.TODO()
)
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type User struct {
	Name   string `redis:"name" mapstructure:"name,omitempty"`
	Age    int64  `redis:"age" mapstructure:"age,omitempty"`
	Email  string `redis:"email"  mapstructure:"email, omitempty"`
	Genres Genres `redis:"genres" mapstructure:"genres, omitempty"`
}

type Genres []string
type Runtime int64

func (g Genres) MarshalBinary() ([]byte, error) {
	values := ""

	for i, val := range g {
		if i == len(g)-1 {
			values = fmt.Sprint(values, val)
		} else {
			values = fmt.Sprint(values, val, ",")
		}
	}
	fmt.Println(values)
	return []byte(strconv.Quote(values)), nil
}

func serializeGeners(g []string) string {
	values := ""

	for i, val := range g {
		if i == len(g)-1 {
			values = fmt.Sprint(values, val)
		} else {
			values = fmt.Sprint(values, val, ",")
		}
	}

	return values
}
func deserializeGenres(val string) []string {

	return strings.Split(val, ",")
}

func main() {

	mail := mailer.New(
		"smtp.sendgrid.net",
		587,
		"apikey",
		"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",
		"auth@greenlight.com")

	data := map[string]interface{}{
		"activationToken": "zsfnzdkfjnksanaksnfasn",
		"userID":          1,
	}
	err := mail.Send("islamghany3@gmail.com", "user_welcome.tmpl", data)
	if err != nil {
		fmt.Println(err)
	}
	// //	keys := []string{"color", "message"}
	// rdb, _ := openRedis("redis-19934.c245.us-east-1-3.ec2.cloud.redislabs.com", "19934", "", "dqFbTrJAXOfhaPaEdHZti8R3HRwcGdx7")

	// defer rdb.Close()
	// // u := User{
	// // 	Name:   "ahmed",
	// // 	Age:    11,
	// // 	Email:  "hello",
	// // 	Genres: []string{"islam", "mostafa"},
	// // }
	// // var m map[string]interface{}
	// // mapstructure.Decode(u, &m)
	// // //m["runtime"] = fmt.Sprintf("%d mins", u.Runtime)
	// // err := Set(rdb, "user4", m, 5*time.Minute)
	// // if err != nil {
	// // 	fmt.Println(err)
	// // }

	// user, err := Get(rdb, "user4")

	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(user)
}

func Set(rdb *redis.Client, key string, values map[string]interface{}, ttl time.Duration) error {
	err := rdb.HSet(ctx, key, values).Err()
	if err != nil {
		return err
	}

	err = rdb.Expire(ctx, key, ttl).Err()
	if err != nil {
		return err
	}
	return nil
}

func Get(rdb *redis.Client, key string) (*User, error) {
	var user User

	data, err := rdb.HGetAll(ctx, key).Result()
	var m map[string]interface{}
	mapstructure.Decode(data, &m)
	m["genres"] = deserializeGenres(data["genres"])
	mapstructure.Decode(m, &user)

	if err != nil {
		return nil, err
	}
	return &user, err
}

// type Movie struct {
// 	ID        int64     `json:"id"`
// 	CreatedAt time.Time `json:"created_at"`
// 	Title     string    `json:"title"`
// 	Year      int32     `json:"year,omitempty"`
// 	Runtime   int       `json:"runtime,omitempty"`
// 	Genres    []string  `json:"genres"`
// 	Version   int32     `json:"version"`
// 	count     int
// }

// func main() {

// 	db, err := openDB(os.Getenv("GREENLIGHT_DB_DSN"))
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	query := `
// 		SELECT  id, created_at, title, year, runtime, genres, version, count
// 		FROM movies
// 		inner join views on views.movie_id = id
// 		WHERE id = $1;
// 	`
// 	var movie Movie

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	err = db.QueryRowContext(ctx, query, 1).Scan(
// 		&movie.ID,
// 		&movie.CreatedAt,
// 		&movie.Title,
// 		&movie.Year,
// 		&movie.Runtime,
// 		pq.Array(&movie.Genres),
// 		&movie.Version,
// 		&movie.count,
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// query = `
// 	// UPDATE views
// 	// SET count = count+1
// 	// WHERE movie_id = $1;
// 	// `
// 	// db.QueryRowContext(ctx, query, 1)
// 	log.Println(movie)
// 	// query := `
// 	// 	INSERT INTO movies (title, year, runtime, genres)
// 	// 	VALUES ($1,$2,$3,$4)
// 	// 	RETURNING id, created_at, version;
// 	// `
// 	// var a [2]string
// 	// a[0] = "islam"
// 	// a[1] = "ahmed"
// 	// args := []interface{}{"new york", 1999, 1885, pq.Array(a)}
// 	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	// defer cancel()
// 	// var movie struct {
// 	// 	ID        int64
// 	// 	CreatedAt time.Time
// 	// 	Version   int32
// 	// }
// 	// err = db.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	// query = `
// 	// 	INSERT INTO views (movie_id)
// 	// 	VALUES ($1);
// 	// `
// 	// err = db.QueryRowContext(ctx, query, movie.ID).Err()
// 	// if err != nil {
// 	// 	log.Fatal(err)
// 	// }
// 	log.Println("all things fine")
// }

// func openDB(dsn string) (*sql.DB, error) {
// 	db, err := sql.Open("postgres", dsn)

// 	if err != nil {
// 		return nil, err
// 	}

// 	// // Set the maximum number of open (in-use + idle) connections in the pool. Note that
// 	// // passing a value less than or equal to 0 will mean there is no limit.
// 	// db.SetMaxOpenConns(conf.db.maxOpenConns)

// 	// // Set the maximum number of idle connections in the pool. Again, passing a value
// 	// // less than or equal to 0 will mean there is no limit.
// 	// db.SetMaxIdleConns(conf.db.maxIdleConns)

// 	// // Use the time.ParseDuration() function to convert the idle timeout duration string
// 	// // to a time.Duration type.
// 	// duration, err := time.ParseDuration(conf.db.maxIdleTime)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// // Set the maximum idle timeout.
// 	// db.SetConnMaxIdleTime(duration)

// 	// Create a context with a 5-second timeout deadline.
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	// Use PingContext() to establish a new connection to the database, passing in the
// 	// context we created above as a parameter. If the connection couldn't be
// 	// established successfully within the 5 second deadline, then this will return an
// 	// error.
// 	err = db.PingContext(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Return the sql.DB connection pool.
// 	return db, nil
// }

func openRedis(host, port, username, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Password: password,
		Addr:     fmt.Sprint(host, ":", port),
		//Username: username,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}
	return rdb, nil
}
