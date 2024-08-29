package models

type User struct {
	ID        uint64
	Email     string
	PassHash  []byte
	FirstName string
	LastName  string
	Age       int32
}
