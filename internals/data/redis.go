package data

import "fmt"

func UsersKey(id int64) string {
	return fmt.Sprint("users#", id)
}
