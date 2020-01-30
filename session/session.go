package session

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jansemmelink/log"
)

//ISession ...
type ISession interface {
	User() IUser
	UserEmail() string
	Expired() bool
	Expiry() time.Time
	Auth(email string, password string) error
	Logout()
	Save() error
}

//Get is called for each HTTP request to get/create a session
func Get(res http.ResponseWriter, req *http.Request) ISession {
	s := &session{
		res:     res,
		req:     req,
		timeout: time.Minute * 5, //= expiry extension on each save
	}
	defer func(s *session) {
		log.Debugf("Session:{user:%s,expiry:%v}", s.UserEmail(), s.Expiry())
	}(s)

	var err error
	s.cookie, err = cookieEncryptor.Get(req, sessionCookieName)
	if err != nil {
		log.Errorf("Failed to get session cookie")
		return s
	}

	log.Tracef("Session cookie: %+v", s.cookie.Values)
	{
		var ok bool
		email, ok := s.cookie.Values["email"].(string)
		if !ok || len(email) == 0 {
			log.Tracef("session cookie.email not defined")
			return s
		}
		s.user = User(email)
		if s.user == nil {
			log.Debugf("session cookie.email=%s is unknown user", email)
			return s
		}
	} //scope for user

	if timeTextString, ok := s.cookie.Values["expiry"].(string); ok {
		if err := s.expiry.UnmarshalText([]byte(timeTextString)); err != nil {
			s.expiry = time.Now().Add(-time.Second)
			log.Errorf("Failed to unmarshal %s (using default: %v): %v", timeTextString, s.expiry, err)
		} else {
			log.Debugf("  Expiry parsed to %v", s.expiry)
		}
	} else {
		log.Errorf("Did not get session expiry from cookie")
	}
	return s
} //Start()

type session struct {
	res     http.ResponseWriter
	req     *http.Request
	timeout time.Duration
	user    IUser
	expiry  time.Time
	cookie  *sessions.Session
}

func (s session) Expired() bool {
	return s.expiry.Before(time.Now())
}

func (s session) Expiry() time.Time {
	return s.expiry
}

func (s session) User() IUser {
	return s.user
}

func (s session) UserEmail() string {
	if s.user == nil {
		return ""
	}
	return s.user.Email()
}

func (s *session) Auth(email string, password string) error {
	u := User(email)
	if u == nil {
		return fmt.Errorf("unknown user")
	}
	if !u.Auth(password) {
		return fmt.Errorf("wrong password")
	}
	s.user = u
	s.expiry = time.Now().Add(time.Minute * 5)
	log.Debugf("  Expiry set to %v", s.expiry)

	//todo: check auth and session id etc against own db
	return nil
}

func (s *session) Logout() {
	s.user = nil
	s.expiry = time.Now().Add(-time.Second)
	log.Debugf("  Expiry set to %v", s.expiry)
}

func (s *session) Save() (err error) {
	defer func() {
		if err != nil {
			log.Errorf("Session not saved: %v", err)
		}
	}()

	if s.cookie == nil {
		s.cookie, err = cookieEncryptor.New(s.req, sessionCookieName)
		if err != nil {
			err = fmt.Errorf("failed to create session cookie: %v", err)
			return
		}
	}

	//when expired (e.g. logged out), do not update the expiry time :-)
	if s.expiry.After(time.Now()) {
		//not expired: extend session expiry on each update before expiry
		s.expiry = time.Now().Add(s.timeout)
		log.Debugf("  Expiry set to %v", s.expiry)
	}
	timeTextBytes, _ := s.expiry.MarshalText()
	s.cookie.Values["expiry"] = string(timeTextBytes)

	if s.user != nil {
		//never clear email - this is used to populate login form
		s.cookie.Values["email"] = s.user.Email()
	}

	if err = s.cookie.Save(s.req, s.res); err != nil { //this encrypts all session data and add it to the res.Header
		err = fmt.Errorf("Failed to save session: %+v", err)
		return
	}
	log.Debugf("Saved session: %+v", s.cookie.Values)
	err = nil
	return
} //session.Save()

//cookie store is used to store data on client systems
//using HTTP cookies. It wraps around the existing HTTP
//methods to store multiple values in a single encrypted
//field on the client system
var cookieEncryptor = sessions.NewCookieStore([]byte("something-very-secret"))

const sessionCookieName = "entry-made-easy-session"
const userdataCookieName = "entry-made-easy-userdata"
