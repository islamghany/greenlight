package data

import "fmt"

// ************** USERS

func UsersKey(id int64) string {
	return fmt.Sprint("users#", id)
}
