package repository

import "strconv"

func nextStringID(nextID *int) string {
	id := strconv.Itoa(*nextID)
	(*nextID)++
	return id
}
