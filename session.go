package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jansemmelink/enter1/page"
	"github.com/jansemmelink/enter1/session"
	"github.com/jansemmelink/log"
)

//info stored in cookies for a session
type sessionInfo struct {
	email  string
	expiry time.Time
}

// func loggedin(res http.ResponseWriter, req *http.Request) bool {
// 	log.Debugf("check if need to login...")

// 	//look for any session related to this request.
// 	session, err := cookieEncryptor.Get(req, sessionCookieName)
// 	if err != nil {
// 		log.Debugf("Failed to get session")
// 		http.Error(res, err.Error(), http.StatusInternalServerError)
// 		return false //not logged in, don't show more content
// 	}

// 	log.Debugf("Got session: %+v", session.Values)

// 	//see when session is valid and not expired
// 	sessionInfo := sessionInfo{}
// 	var ok bool
// 	if sessionInfo.email, ok = session.Values["email"].(string); !ok || len(sessionInfo.email) == 0 {
// 		log.Debugf("session.email not defined")
// 		showLoginForm(res, req)
// 		return false
// 	}

// 	if timeTextString, ok := session.Values["expiry"].(string); ok {
// 		if err := sessionInfo.expiry.UnmarshalText([]byte(timeTextString)); err == nil {
// 			log.Errorf("Failed to unmarshal %s into time: %v", timeTextString, err)
// 		}
// 	} else {
// 		log.Errorf("Did not get session expiry from cookie")
// 	}
// 	if sessionInfo.expiry.Before(time.Now()) {
// 		log.Debugf("session expired")
// 		showLoginForm(res, req)
// 		return false
// 	}

// 	//todo: check auth and session id etc against own db

// 	//has unexpired session: user is loggedin
// 	return true
// }

func showRegisterForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Start(res, "Registreer")
	t := page.LoadTmpl("register-form")
	t.Render(res, req, nil)
	page.End(res)
}

func authRegister(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Error(res, req, fmt.Errorf("NYI"))
}

func showLoginForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Start(res, "Teken In")
	t := page.LoadTmpl("login-form")
	t.Render(res, req, map[string]interface{}{"email": s.UserEmail()})
	page.End(res)
}

func authLogin(res http.ResponseWriter, req *http.Request, s session.ISession) {
	req.ParseForm()
	email := req.FormValue("email")
	password := req.FormValue("password")
	if err := s.Auth(email, password); err != nil {
		log.Debugf("Auth failed: %v", err)
		showLoginForm(res, req, s)
		return
	}

	//todo: redirect to the page originally requested...
	//todo: not sure yet how to get that, could put in session...?
	//for now, always redirect to entries list
	// log.Debugf("Redirect to /entries...")
	// http.Redirect(res, req, "/entries", http.StatusTemporaryRedirect)

	//rather just confirm for now... todo: try redirect again when cookies work...
	s.Save()
	page.Start(res, "Logged in")
	res.Write([]byte("You are now logged in. Go to <a href=\"/entries\">Entries</a> ..."))
	page.End(res)
}

func showResetForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Start(res, "Wagwoord")
	t := page.LoadTmpl("reset-form")
	t.Render(res, req, nil)
	page.End(res)
}

func authReset(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Error(res, req, fmt.Errorf("NYI"))
}

func showChangeForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Start(res, "Wagwoord")
	t := page.LoadTmpl("change-form")
	t.Render(res, req, nil)
	page.End(res)
}

func authChange(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Error(res, req, fmt.Errorf("NYI"))
}

func authLogout(res http.ResponseWriter, req *http.Request, s session.ISession) {
	s.Logout()
	s.Save()
	page.Start(res, "Totsiens")
	res.Write([]byte(`<h1>Totsiens</h1><p>Jy is nou uitgeteken.</p>`))
	page.End(res)
}

//use secure around handlers that need to have a valid authenticated session
//the handler must save session data if it changed, before writing the response
func secure(handler func(res http.ResponseWriter, req *http.Request, s session.ISession)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Debugf("HTTP %s %s", req.Method, req.URL.Path)
		s := session.Get(res, req)
		log.Debugf("  expiry:  %v", s.Expiry())
		log.Debugf("  expired: %v", s.Expired())
		log.Debugf("  user:    %v", s.User())
		if s.Expired() || s.User() == nil {
			showLoginForm(res, req, s) //todo redirect to requests page after login
			return
		}

		handler(res, req, s)
	}
} //secure()

//use oper around handlers that does not require authenticated session
func open(handler func(res http.ResponseWriter, req *http.Request, s session.ISession)) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		log.Debugf("HTTP %s %s", req.Method, req.URL.Path)
		s := session.Get(res, req)
		handler(res, req, s)
	}
} //secure()
