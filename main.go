package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/pat"
	"github.com/jansemmelink/enter1/org"
	"github.com/jansemmelink/enter1/page"
	"github.com/jansemmelink/enter1/person"
	"github.com/jansemmelink/enter1/session"
	"github.com/jansemmelink/log"
	"github.com/satori/uuid"
)

func main() {
	log.DebugOn()
	http.ListenAndServe("localhost:8000", app())
}

func app() http.Handler {
	r := pat.New()
	r.Get("/auth/register", open(showRegisterForm))
	r.Post("/auth/register", open(authRegister))
	r.Get("/auth/login", open(showLoginForm))
	r.Post("/auth/login", open(authLogin))
	r.Get("/auth/reset", open(showResetForm))
	r.Post("/auth/reset", open(authReset))
	r.Get("/auth/change", open(showChangeForm))
	r.Post("/auth/change", open(authChange))
	r.Get("/auth/logout", open(authLogout))

	//open links - auth not required:
	r.Get("/enter", enterForm)
	r.Post("/enter", enterHandler)
	r.Get("/entry/{id}", entryView)

	r.Get("/entries", secure(listOfEntries))
	r.Get("/delete/{id}", secure(entryDelete))

	//organisations
	r.Get("/orgs", open(org.ShowList))
	r.Get("/org/new", open(org.NewOrgForm))
	r.Post("/org/new", open(org.NewOrgPost))
	r.Get("/org/view/{id}", open(org.ShowOrg))
	//r.Get("/org/edit/{id}", secure(org.OrgEditForm))
	//r.Get("/org/delete/{id}", secure(org.OrgDelete))

	//people
	r.Get("/persons", secure(person.ShowList))
	r.Get("/person/new", open(person.NewPersonForm))
	r.Post("/person/new", open(person.NewPersonPost))
	r.Get("/person/view/{id}", open(person.ShowPerson))
	//r.Get("/person/edit/{id}", secure(person.PersonEditForm))
	//r.Get("/person/delete/{id}", secure(person.PersonDelete))

	//defaults:
	r.Get("/resource", resourceHandler)
	r.Get("/", unknownHandler)
	return r
}

//resourceHandler is called to serve resource files such as css, fonts, js, ...
func resourceHandler(res http.ResponseWriter, req *http.Request) {
	filename := "./assets" + req.URL.Path[9:] //skip: /resource
	log.Debugf("Serving file: %s\n", filename)
	http.ServeFile(res, req, filename)
}

//unknownHandler is called for "/" or any unknown URL path
func unknownHandler(res http.ResponseWriter, req *http.Request) {
	//without any path, default to home page to show the form
	if req.URL.Path == "/" {
		enterForm(res, req)
		return
	}

	//unknown URL, show an error
	page.Start(res, "Error")
	res.Write([]byte(`
	<h1>Error</h1>
	<p>The link you entered does not exist on this system.</p>
	<p>Click <a href="/enter">here</a> to go home.</p>
	`))
	page.End(res)
}

//enterForm shows the form for a new entry
//it is used as home page
func enterForm(res http.ResponseWriter, req *http.Request) {
	log.Debugf("Showing entry form\n")
	page.Start(res, "Inskrywingsvorm")
	t := page.LoadTmpl("form-enter")
	t.Render(res, req, nil)
	page.End(res)
}

//enterHandler processes the submitted form data to create a new entry
func enterHandler(res http.ResponseWriter, req *http.Request) {
	log.Debugf("Entry submitted\n")
	if err := req.ParseForm(); err != nil {
		page.Error(res, req, err)
		log.Debugf("Submit form parsing failed: %+v\n", err)
		return
	}

	//form data example:
	//	map[dob:[4] email:[a@b.c] firstname:[2] gender:[female] lastname:[3]]
	e := entry{
		FirstName: req.FormValue("firstname"),
		LastName:  req.FormValue("lastname"),
		Dob:       req.FormValue("dob"),
		Gender:    req.FormValue("gender"),
		Email:     req.FormValue("email"),
		Phone:     req.FormValue("phone"),
	}
	log.Debugf("Entry: %+v\n", e)
	if err := e.Validate(); err != nil {
		page.Error(res, req, fmt.Errorf("invalid entry: %v", err))
		log.Debugf("Invalid entry: %+v\n", err)
		return
	}

	e.id = uuid.NewV1().String()
	//...entry.Validate()
	if err := e.Save(); err != nil {
		page.Error(res, req, err)
		log.Debugf("Entry failed: %+v\n", err)
		return
	}

	page.Start(res, "Entry Submitted")
	res.Write([]byte(`
	<h1>Thank you</h1>
	<p>Your reference is: ` + e.id + `</p>
	<p>You can view your entry with this <a href="http://localhost:8000/entry/` + e.id + `">link</a>.</p>
	`))
	page.End(res)
}

func entryView(res http.ResponseWriter, req *http.Request) {
	e := entry{}
	e.id = req.URL.Query().Get(":id")
	log.Debugf("Loading id=%s", e.id)
	if err := e.Load(); err != nil {
		page.Error(res, req, err)
		log.Debugf("Failed to load id=%s: %+v\n", e.id, err)
		return
	}

	page.Start(res, "Your Entry")
	res.Write([]byte(`
	<h1>Entry: ` + e.id + `</h1>
	<p>Name: ` + e.FirstName + `</p>
	<p>Surname: ` + e.LastName + `</p>
	<p>Gender: ` + e.Gender + `</p>
	<p>Date of birth: ` + e.Dob + `</p>
	<a href="/edit/` + e.id + `">Edit</a>
	<a href="/delete/` + e.id + `">Delete</a>
	`))
	page.End(res)
}

func entryDelete(res http.ResponseWriter, req *http.Request, s session.ISession) {
	e := entry{}
	e.id = req.URL.Query().Get(":id")
	log.Debugf("Loading id=%s", e.id)
	if err := e.Load(); err != nil {
		page.Error(res, req, err)
		log.Debugf("Failed to load id=%s: %+v\n", e.id, err)
		return
	}
	os.Remove(e.Filename())
	http.Redirect(res, req, "/entries", http.StatusTemporaryRedirect)
}

func listOfEntries(res http.ResponseWriter, req *http.Request, s session.ISession) {
	//get list of all entries
	list, err := loadList()
	if err != nil {
		page.Error(res, req, fmt.Errorf("failed to load the list: %v", err))
		return
	}

	//filter list

	//show list
	page.Start(res, "Entries")
	size := intParam(req.URL.Query().Get("size"), 10)
	res.Write([]byte(`<div class="container">`))
	res.Write([]byte(`<h1>Entries</h1>
	<table class="table">
	<thead>
	<tr><th scope="col">Name</th><th scope="col">Gender</th><th scope="col">Dob</th><th scope="col">Manage</th></tr>
	</thead>
	<tbody>`))
	for i, e := range list {
		if i > size {
			break
		}
		res.Write([]byte(fmt.Sprintf(`<tr><td><a href="%s">%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			"/entry/"+e.id,
			e.FirstName+" "+e.LastName,
			e.Gender,
			e.Dob,
			"<a href=\"/delete/"+e.id+"\">Delete</a>"))) //actions
	}
	res.Write([]byte(`
	</tbody>
	</table>
	</div>`))
	page.End(res)
}

/*
NEXT:
- commit...
- store group info in db...
- add member to a org:
	- general invite for anyone to fill in application form
	- store personal info in db (not user, just a person)
	- after login - show person info from email
	- later: send invite by email to register that email only...
	- list of applications: accept/reject -> send email to next to upload POP
	- upload pop
	- list of pop - accept -> confirmed member
	- send invite to all members of a existing group

- access control - need to store id and ask password on new session - not needed for form
- access control - only admin/owner allowed to edit entries
- vorm ingee moet epos stuur met skakel
- form validation with feedback to the user
- hersien vorm en stuur terug met epos om te redigeer en hersien weer
- lys van vorms wat hersien moet word
- toegang beheer op hersiening
- delete duplikaat entries (toegang beheer)
- laai bewys van betaling teen 'n vorm
- lys van betaalde / onbetaalde vorms
- hersien bewys van betaling + aanvaar/reject

- improve:
- make unit of "entry" with struct, form template and validation then make other types of items

DONE:
- basic server with form, validate, save and load
- added org and person with parent/child relationships
*/
