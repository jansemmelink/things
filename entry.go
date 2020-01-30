package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jansemmelink/log"
)

type entry struct {
	id string
	//personal details:
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Dob       string `json:"dob"`
	Gender    string `json:"gender"`
	//contact info
	Email string `json:"email"`
	Phone string `json:"phone"`
}

func (e entry) Validate() error {
	if len(e.FirstName) < 1 {
		return fmt.Errorf("no name")
	}
	return nil
}

func (e entry) Save() error {
	f, err := os.Create(e.Filename())
	if err != nil {
		return fmt.Errorf("cannot create entry file %s: %v", e.Filename(), err)
	}
	defer f.Close()
	jsonEntry, _ := json.Marshal(e)
	if _, err := f.Write(jsonEntry); err != nil {
		return fmt.Errorf("cannot write to entry file %s", e.Filename())
	}
	return nil
}

func (e *entry) Load() error {
	log.Debugf("Load(%+v)", e)
	f, err := os.Open(e.Filename())
	if err != nil {
		return fmt.Errorf("cannot open entry file %s: %v", e.Filename(), err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&e)
	if err != nil {
		return fmt.Errorf("invalid JSON in entry file: %v", err)
	}
	//e.Validate()
	log.Debugf("loaded(%+v)", e)
	return nil
}

const entryDir = "./entries"

func (e entry) Filename() string {
	return entryDir + "/" + e.id + ".json"
}

func loadList() ([]entry, error) {
	list := []entry{}
	if filepath.Walk(
		"./entries",
		func(fn string, info os.FileInfo, err error) error {
			log.Debugf("walk: %s", fn)
			if info.Mode().IsRegular() && strings.HasSuffix(fn, ".json") {
				id := strings.TrimRight(path.Base(fn), ".json")
				e := entry{id: id}
				if err := e.Load(); err != nil {
					log.Errorf("failed to load %s: %v", id, err)
				} else {
					log.Debugf("loaded: %+v", e)
					list = append(list, e)
				}
			}
			return nil
		}) != nil {
		return nil, fmt.Errorf("failed to load the list")
	}
	return list, nil
}
