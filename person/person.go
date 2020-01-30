package person

import (
	"fmt"
	"net/http"

	"github.com/jansemmelink/enter1/page"
	"github.com/jansemmelink/enter1/session"
	"github.com/jansemmelink/log"
	"github.com/satori/uuid"
)

//ShowList ...
func ShowList(res http.ResponseWriter, req *http.Request, s session.ISession) {
	page.Start(res, "Mense")
	res.Write([]byte(`<div class="container">`))
	res.Write([]byte(`<h1>Mense</h1>
	<ul>`))
	for _, person := range allPersonsByID {
		res.Write([]byte(fmt.Sprintf(`<li><a href="/person/view/%s">%s</a></li>`, person.id, person.Name())))
	}
	res.Write([]byte(`</ul>`))
	res.Write([]byte(`<a href="/person/new">Skep</a>`))
	res.Write([]byte(`</div>`))
	page.End(res)
}

//NewPersonForm ...
func NewPersonForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	data := map[string]string{}
	parentID := req.URL.Query().Get("parent")
	if len(parentID) > 0 {
		parentPerson, ok := allPersonsByID[parentID]
		if !ok {
			page.Error(res, req, fmt.Errorf("unknown parent id: %s", parentID))
			return
		}
		data["parentID"] = parentID
		data["parentName"] = parentPerson.Name()
	}

	page.Start(res, "Nuwe Persoon")
	t := page.LoadTmpl("person/person-form")
	t.Render(res, req, data)
	page.End(res)
}

//NewPersonPost ...
func NewPersonPost(res http.ResponseWriter, req *http.Request, s session.ISession) {
	parentID := req.URL.Query().Get("parent")
	var parentPerson *person
	if len(parentID) > 0 {
		var ok bool
		parentPerson, ok = allPersonsByID[parentID]
		if !ok {
			page.Error(res, req, fmt.Errorf("unknown parent id: %s", parentID))
		}
	}
	req.ParseForm()

	newPerson := &person{
		id:         uuid.NewV1().String(),
		parents:    []*person{},
		ParentIDs:  []string{},
		FirstName:  req.FormValue("firstname"),
		OtherNames: req.FormValue("othernames"),
		LastName:   req.FormValue("lastname"),
		Dob:        req.FormValue("dob"),
		Gender:     req.FormValue("gender"),
		//Nationality string `json:"nationality"`
		//NationalID string `json:"national_id"`
		//contact info
		Email:    req.FormValue("email"),
		Phone:    req.FormValue("phone"),
		AltPhone: req.FormValue("alt_phone"),
	}

	if parentPerson != nil {
		newPerson.parents = append(newPerson.parents, parentPerson)
		newPerson.ParentIDs = append(newPerson.ParentIDs, parentPerson.id)
	}
	//todo: ability to add more parents ...
	//todo: ability to create child and to create a new parent
	//todo: add family groups: no strict parent-child, just all in same group

	//create the new org
	allPersonsByID[newPerson.id] = newPerson
	showPerson(res, req, s, newPerson)
}

//ShowPerson ...
func ShowPerson(res http.ResponseWriter, req *http.Request, s session.ISession) {
	log.Debugf("Showing person.id=%s", req.URL.Query().Get(":id"))
	p, ok := allPersonsByID[req.URL.Query().Get(":id")]
	if !ok {
		page.Error(res, req, fmt.Errorf("Unknown person ID"))
		return
	}
	showPerson(res, req, s, p)
}

func showPerson(res http.ResponseWriter, req *http.Request, s session.ISession, p *person) {
	//data for template
	data := map[string]interface{}{}
	data["id"] = p.id
	data["name"] = p.Name()
	data["firstname"] = p.FirstName
	data["othernames"] = p.OtherNames
	data["lastname"] = p.LastName
	data["dob"] = p.Dob
	data["gender"] = p.Gender
	data["email"] = p.Email
	data["phone"] = p.Phone
	data["alt_phone"] = p.AltPhone
	//data["nationality"] = p.Nationality
	//data["national_id"] = p.NationalID
	if p.parents != nil && len(p.parents) > 0 {
		parentsData := []map[string]interface{}{}
		for _, pp := range p.parents {
			parentData := map[string]interface{}{}
			parentData["name"] = pp.Name()
			parentData["id"] = pp.id
			parentsData = append(parentsData, parentData)
		}
		data["parents"] = parentsData
	}
	children := p.Children()
	childrenData := []map[string]interface{}{}
	for _, child := range children {
		childData := map[string]interface{}{}
		childData["id"] = child.id
		childData["name"] = child.Name()
		childrenData = append(childrenData, childData)
	}
	data["children"] = childrenData

	//show template
	page.Start(res, "Person: "+p.Name())
	t := page.LoadTmpl("person/person-view")
	t.Render(res, req, data)

	//management links: todo: only if owner + get delete confirm!!!
	res.Write([]byte(`
	<a href="/person/edit/` + p.id + `">Edit</a>
	<a href="/person/delete/` + p.id + `">Delete</a>
	`))
	page.End(res)
}

var (
	allPersonsByID = map[string]*person{}
)

type person struct {
	//runtime
	id       string
	parents  []*person
	children []*person
	//in db
	ParentIDs  []string `json:"parent_ids"`
	FirstName  string   `json:"firstname"`
	OtherNames string   `json:"othernames"`
	LastName   string   `json:"lastname"`
	Dob        string   `json:"dob"`
	Gender     string   `json:"gender"`
	//Nationality string `json:"nationality"`
	//NationalID string `json:"national_id"`
	//contact info
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	AltPhone string `json:"alt_phone"`
}

func (p person) Name() string {
	return p.FirstName + " " + p.LastName
}

func (p person) Children() []*person {
	children := []*person{}
	for _, other := range allPersonsByID {
		//check if o is one of the parents of this person
		for _, otherParentID := range other.ParentIDs {
			if p.id == otherParentID {
				//o is a parent of this p
				children = append(children, other)
			}
		}
	}
	return children
}
