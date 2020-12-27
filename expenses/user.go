package expenses

type User struct {
	ID       uint
	Email    string
	Password string
}

type CreateUserRequest struct {
	Email    string
	Password string
}

type UserReponse struct {
	ID    uint
	Email string
}
