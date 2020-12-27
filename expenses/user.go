package expenses

type User struct {
	ID       uint
	Email    string
	Password string
}

type CreateUser struct {
	Email    string
	Password string
}
