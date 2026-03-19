// Package pager provides pagination calculation helpers.
package pager

import (
	"net/http"
	"strconv"
)

// Pager holds pagination state for a single page of results.
type Pager struct {
	Page       int
	PerPage    int
	TotalItems int
	TotalPages int
}

// New reads page and per_page query params from r, clamping page to ≥ 1 and
// using defaultPerPage when the query param is absent or invalid.
func New(r *http.Request, defaultPerPage int) Pager {
	page := 1
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	perPage := defaultPerPage
	if pp, err := strconv.Atoi(r.URL.Query().Get("per_page")); err == nil && pp > 0 {
		perPage = pp
	}
	return Pager{Page: page, PerPage: perPage}
}

// SetTotal updates TotalItems and computes TotalPages.
func (p *Pager) SetTotal(total int) {
	p.TotalItems = total
	if p.PerPage <= 0 {
		p.TotalPages = 0
		return
	}
	p.TotalPages = (total + p.PerPage - 1) / p.PerPage
}

// Offset returns the SQL OFFSET value for the current page.
func (p *Pager) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

func (p *Pager) IsFirst() bool { return p.Page <= 1 }

func (p *Pager) IsLast() bool { return p.Page >= p.TotalPages }

func (p *Pager) PrevPage() int {
	if p.Page <= 1 {
		return 1
	}
	return p.Page - 1
}

func (p *Pager) NextPage() int {
	if p.Page >= p.TotalPages {
		return p.TotalPages
	}
	return p.Page + 1
}
