package db

import (
	// "core.vm"
	"encoding/json"
	"sync"

	"github.com/blevesearch/bleve"
)

// DB the database object
type DB struct {
	index      bleve.Index
	globals    map[string]interface{}
	gl         sync.RWMutex
	crons      map[string]*Cron
	cl         sync.RWMutex
	CronReload chan int
}

// Open opens the specified database
func Open(dbname string) (*DB, error) {
	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dbname, mapping)
	if err != nil && err == bleve.ErrorIndexPathExists {
		index, err = bleve.Open(dbname)
	}
	if err != nil {
		return nil, err
	}
	db := &DB{
		index: index,
		gl:    sync.RWMutex{},
	}
	db.globals = db.GlobalsLoad()
	db.crons = db.CronsLoad()
	db.cl = sync.RWMutex{}
	db.CronReload = make(chan int)
	return db, nil
}

// Put insert the specified Procedure
func (db *DB) Put(p *Procedure) error {
	for i, t := range p.Tags {
		if t == "" {
			p.Tags = append(p.Tags[0:i], p.Tags[i+1:]...)
		}
	}
	return db.index.Index(p.Key, p)
}

// Delete delete the specified procedure
func (db *DB) Delete(key string) error {
	return db.index.Delete(key)
}

// Find search for a procedure
func (db *DB) Find(q string, sortby []string, offset, limit int) (*Result, error) {
	searchQuery := bleve.NewQueryStringQuery(q)
	searchRequest := bleve.NewSearchRequest(searchQuery)
	searchRequest.Fields = []string{"*"}
	if len(sortby) < 1 {
		sortby = []string{"-_score"}
	}
	searchRequest.SortBy(sortby)
	if offset >= 0 {
		searchRequest.From = offset
	}
	if limit > 0 {
		searchRequest.Size = limit
	}
	searchResults, err := db.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	hits := []map[string]interface{}{}
	for _, hit := range searchResults.Hits {
		hits = append(hits, hit.Fields)
	}
	return &Result{
		Totals:   searchResults.Total,
		Hits:     hits,
		MaxScore: searchResults.MaxScore,
		Took:     searchResults.Took,
	}, nil
}

// GlobalsSet set global var(s)
func (db *DB) GlobalsSet(data map[string]interface{}) {
	db.gl.Lock()
	defer db.gl.Unlock()
	for k, v := range data {
		db.globals[k] = v
	}
	j, _ := json.Marshal(db.globals)
	db.index.SetInternal([]byte("internals/vars/globals"), j)
}

// GlobalsLoad freshly load the globals
func (db *DB) GlobalsLoad() map[string]interface{} {
	vars := map[string]interface{}{}
	data, _ := db.index.GetInternal([]byte("internals/vars/globals"))
	json.Unmarshal(data, &vars)
	return vars
}

// GlobalsGet get the globals from the cache
func (db *DB) GlobalsGet() map[string]interface{} {
	return db.globals
}

// GlobalsUnset unset the specified key(s) from the globals
func (db *DB) GlobalsUnset(keys []string) {
	vars := db.GlobalsGet()
	for _, k := range keys {
		delete(vars, k)
		delete(db.globals, k)
	}
	db.GlobalsSet(vars)
}

// CronsLoad list all cron jobs
func (db *DB) CronsLoad() map[string]*Cron {
	crons := map[string]*Cron{}
	data, _ := db.index.GetInternal([]byte("internals/crons"))
	json.Unmarshal(data, &crons)
	return crons
}

// CronsSet set a cron-job
func (db *DB) CronsSet(key, interval, fn string) error {
	db.cl.Lock()
	defer db.cl.Unlock()
	defer db.CronsRun()
	db.crons[key] = &Cron{
		Interval: interval,
		Job:      fn,
	}
	j, _ := json.Marshal(db.crons)
	return db.index.SetInternal([]byte("internals/crons"), j)
}

// CronsUnset unset a cron-job
func (db *DB) CronsUnset(key string) error {
	db.cl.Lock()
	defer db.cl.Unlock()
	defer db.CronsRun()
	delete(db.crons, key)
	j, _ := json.Marshal(db.crons)
	return db.index.SetInternal([]byte("internals/crons"), j)
}

// CronsGet get the cached crons
func (db *DB) CronsGet() map[string]*Cron {
	return db.crons
}

// CronsRun runs the crons kernel
func (db *DB) CronsRun() {
	go func() {
		db.CronReload <- 1
	}()
}
