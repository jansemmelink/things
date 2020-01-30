package page

import (
	"html/template"
	"net/http"
	"sync"

	"github.com/jansemmelink/log"
)

//LoadTmpl loads the template: name must exist in ./templates/<name>.html
func LoadTmpl(name string) Template {
	templateMutex.Lock()
	defer templateMutex.Unlock()
	if t, ok := templateByName[name]; ok {
		return t
	}

	tmplFilename := "./templates/" + name + ".html"
	tmpl, err := template.ParseFiles(tmplFilename)
	if err != nil {
		log.Errorf("template error: %s: %v", tmplFilename, err)
		tmpl, _ = template.New(name).Parse(`
		<h1>Error</h1>
		<p>Page not found.</p>`)
	}
	return Template{
		tmpl: tmpl,
	}
}

//Template ...
type Template struct {
	tmpl *template.Template
}

//Render template into http response writer
func (t Template) Render(res http.ResponseWriter, req *http.Request, data interface{}) {
	log.Debugf("Rendering data=%+v", data)
	if err := t.tmpl.Execute(res, data); err != nil {
		log.Errorf("Render Error: %+v", err)
	}
	log.Debugf("Render success")
}

var (
	templateMutex  sync.Mutex
	templateByName = make(map[string]Template)
)
