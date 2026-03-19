package pager

import (
	"net/http/httptest"
	"testing"
)

func TestNewAndSetTotalComputePaginationState(t *testing.T) {
	req := httptest.NewRequest("GET", "/users?page=3&per_page=25", nil)
	p := New(req, 10)
	p.SetTotal(240)

	if p.Page != 3 || p.PerPage != 25 || p.TotalPages != 10 {
		t.Fatalf("Pager state = %#v", p)
	}
	if p.Offset() != 50 || p.PrevPage() != 2 || p.NextPage() != 4 {
		t.Fatalf("Pager navigation = %#v", p)
	}
}

func TestPagerClampsInvalidInput(t *testing.T) {
	req := httptest.NewRequest("GET", "/users?page=0&per_page=-1", nil)
	p := New(req, 20)
	p.SetTotal(0)

	if p.Page != 1 || p.PerPage != 20 || !p.IsFirst() || !p.IsLast() {
		t.Fatalf("Pager invalid input result = %#v", p)
	}
}
