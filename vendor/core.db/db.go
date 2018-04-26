package db

import (
	"github.com/blevesearch/bleve"
)

// DB the database object
type DB struct {
	index bleve.Index
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
	}
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
