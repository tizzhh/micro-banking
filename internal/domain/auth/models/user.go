package models

type User struct {
	ID        uint64
	Email     string
	PassHash  []byte
	FirstName string
	LastName  string
	Balance   uint64 // In USD cents
	Age       uint32
}
