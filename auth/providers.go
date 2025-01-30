package auth

type User struct {
	Name  string
	Email string
}

type Provider struct {
	Url  string
	Name string
}
