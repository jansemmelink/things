package org

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
	page.Start(res, "Organisations")
	res.Write([]byte(`<div class="container">`))
	res.Write([]byte(`<h1>Organisasies</h1>
	<ul>`))
	for _, org := range topsByID { //only top groups
		res.Write([]byte(fmt.Sprintf(`<li><a href="/org/view/%s">%s</a></li>`, org.id, org.Name)))
	}
	res.Write([]byte(`</ul>`))
	res.Write([]byte(`<a href="/org/new">Skep</a>`))
	res.Write([]byte(`</div>`))
	page.End(res)
}

//NewOrgForm ...
func NewOrgForm(res http.ResponseWriter, req *http.Request, s session.ISession) {
	data := map[string]string{}
	parentID := req.URL.Query().Get("parent")
	if len(parentID) > 0 {
		parentOrg, ok := orgsByID[parentID]
		if !ok {
			page.Error(res, req, fmt.Errorf("unknown parent id: %s", parentID))
			return
		}
		data["parentID"] = parentID
		data["parentName"] = parentOrg.Name
	}

	page.Start(res, "Organisations")
	t := page.LoadTmpl("org/org-form")
	t.Render(res, req, data)
	page.End(res)
}

//NewOrgPost ...
func NewOrgPost(res http.ResponseWriter, req *http.Request, s session.ISession) {
	parentID := req.URL.Query().Get("parent")
	var parentOrg *org
	if len(parentID) > 0 {
		var ok bool
		parentOrg, ok = orgsByID[parentID]
		if !ok {
			page.Error(res, req, fmt.Errorf("unknown parent id: %s", parentID))
		}
	}
	req.ParseForm()

	org := &org{
		id:       uuid.NewV1().String(),
		parent:   parentOrg,
		ParentID: parentID,
		Name:     req.FormValue("name"),
		Desc:     req.FormValue("desc"),
	}
	//create the new org
	orgsByID[org.id] = org
	if org.parent == nil {
		topsByID[org.id] = org
	}
	showOrg(res, req, s, org)
}

//ShowOrg ...
func ShowOrg(res http.ResponseWriter, req *http.Request, s session.ISession) {
	log.Debugf("Showing org.id=%s", req.URL.Query().Get(":id"))
	org, ok := orgsByID[req.URL.Query().Get(":id")]
	if !ok {
		page.Error(res, req, fmt.Errorf("Unknown organisation ID"))
		return
	}
	showOrg(res, req, s, org)
}

func showOrg(res http.ResponseWriter, req *http.Request, s session.ISession, org *org) {
	//data for template
	data := map[string]interface{}{}
	data["name"] = org.Name
	data["desc"] = org.Desc
	data["id"] = org.id
	if org.parent != nil {
		data["parentID"] = org.parent.id
		data["parentName"] = org.parent.Name
	}
	subs := org.Subs()
	groupsList := []map[string]interface{}{}
	for _, sub := range subs {
		groupData := map[string]interface{}{}
		groupData["id"] = sub.id
		groupData["name"] = sub.Name
		groupsList = append(groupsList, groupData)
	}
	data["groups"] = groupsList

	//show template
	page.Start(res, "Organisasie: "+org.Name)
	t := page.LoadTmpl("org/org-view")
	t.Render(res, req, data)

	//management links: todo: only if owner + get delete confirm!!!
	res.Write([]byte(`
	<a href="/org/edit/` + org.id + `">Edit</a>
	<a href="/org/delete/` + org.id + `">Delete</a>
	`))
	page.End(res)
}

var (
	orgsByID = map[string]*org{}
	topsByID = map[string]*org{}
)

type org struct {
	//memory only
	id     string
	parent *org //nil if this has no parent

	//stored in db
	ParentID string
	Name     string
	Desc     string
}

func (o org) Subs() []*org {
	subList := []*org{}
	for _, other := range orgsByID {
		if other.ParentID == o.id {
			subList = append(subList, other)
		}
	}
	return subList
}
