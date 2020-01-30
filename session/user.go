package session

//IUser ...
type IUser interface {
	Email() string
	Auth(password string) bool
}

type user struct {
	email string
}

func (u user) Email() string {
	return u.email
}

func (u user) Auth(password string) bool {
	//only allow hard coded admin for now
	if u.email == "jan.semmelink@gmail.com" &&
		password == "" {
		return true
	}
	return false
}

//User gets existing user, else nil
func User(email string) IUser {
	return user{email: email} //todo: load
}

//NewUser creates a user, fail if already exists
func NewUser(email string) IUser {
	return user{email: email} //todo: load
}
