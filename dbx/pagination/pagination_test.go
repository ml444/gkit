package pagination

import "testing"

func TestNewDefaultPagination(t *testing.T) {
	p := NewDefaultPagination()
	if p.Page != 1 || p.Size != 10 || p.Total != 0 || p.SkipCount {
		t.Errorf("NewDefaultPagination() = %+v, want default values", p)
	}
}

func TestSetPage(t *testing.T) {
	p := NewDefaultPagination()
	p.SetPage(5)
	if p.Page != 5 {
		t.Errorf("SetPage(5) = %d, want 5", p.Page)
	}

	p.SetPage(0)
	if p.Page != 1 {
		t.Errorf("SetPage(0) = %d, want 1", p.Page)
	}
}

func TestSetNextPage(t *testing.T) {
	p := NewDefaultPagination()
	p.Total = 100
	p.Size = 10

	ok := p.SetNextPage()
	if !ok || p.Page != 2 {
		t.Errorf("SetNextPage() = %t, %d, want true, 2", ok, p.Page)
	}

	for i := 0; i < 8; i++ {
		p.SetNextPage()
	}
	ok = p.SetNextPage()
	if ok || p.Page != 11 {
		t.Errorf("SetNextPage() = %t, %d, want false, 10", ok, p.Page)
	}
}

func TestOffset(t *testing.T) {
	p := NewDefaultPagination()
	if p.Offset() != 0 {
		t.Errorf("Offset() = %d, want 0", p.Offset())
	}

	p.SetPage(2)
	p.SetSize(20)
	if p.Offset() != 20 {
		t.Errorf("Offset() = %d, want 20", p.Offset())
	}
}
