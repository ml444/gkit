package pagination

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestNewDefaultPagination(t *testing.T) {
	p := NewDefaultPagination()
	if p.Page != 1 || p.Size != 10  || p.SkipCount {
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


func TestPaginationMutatorsAndOffset(t *testing.T) {
	p := NewDefaultPagination()
	if p.SetPage(0) != p || p.Page != 1 {
		t.Fatal("SetPage should normalize zero and chain")
	}
	if p.SetPageAndSize(3, 20) != p || p.Page != 3 || p.Size != 20 {
		t.Fatalf("SetPageAndSize = %+v", p)
	}
	if p.SetSize(30) != p || p.Size != 30 {
		t.Fatalf("SetSize = %+v", p)
	}
	p.SetSize(20)
	if p.SetSkipCount() != p || !p.SkipCount {
		t.Fatal("SetSkipCount should chain and set flag")
	}
	if p.Offset() != 40 {
		t.Errorf("Offset() = %d, want 40", p.Offset())
	}
	p.SetPage(1)
	if p.Offset() != 0 {
		t.Errorf("Offset page 1 = %d", p.Offset())
	}
}


func TestScrollMutators(t *testing.T) {
	s := (&Scroll{}).SetSize(30).SetCursor("cursor-1")
	if s.Size != 30 || s.Cursor != "cursor-1" {
		t.Fatalf("Scroll = %+v", s)
	}
}

func TestGeneratedProtoMethods(t *testing.T) {
	cases := []struct {
		name       string
		msg        protoreflect.ProtoMessage
		reset      func()
		descriptor func() ([]byte, []int)
		check      func()
		nilCheck   func()
	}{
		{
			name: "pagination",
			msg:  &Pagination{Page: 2, Size: 3, SkipCount: true},
			reset: func() {
				p := &Pagination{Page: 2, Size: 3, SkipCount: true}
				p.Reset()
				if p.GetPage() != 0 || p.GetSize() != 0 || p.GetSkipCount() {
					t.Fatalf("pagination reset = %+v", p)
				}
			},
			descriptor: (&Pagination{}).Descriptor,
			check: func() {
				p := &Pagination{Page: 2, Size: 3, SkipCount: true}
				if p.String() == "" || p.ProtoReflect().Descriptor().FullName() != "pagination.Pagination" {
					t.Fatal("pagination proto reflection failed")
				}
				p.ProtoMessage()
				if p.GetPage() != 2 || p.GetSize() != 3 || !p.GetSkipCount() {
					t.Fatalf("pagination getters = %+v", p)
				}
			},
			nilCheck: func() {
				var p *Pagination
				if p.ProtoReflect().Descriptor().FullName() != "pagination.Pagination" || p.GetPage() != 0 || p.GetSize() != 0 || p.GetSkipCount() {
					t.Fatal("nil pagination getters/reflection failed")
				}
			},
		},
		{
			name: "pagination_result",
			msg:  &PaginationResult{Page: 2, Size: 3, Total: 4, SkipCount: true},
			reset: func() {
				p := &PaginationResult{Page: 2, Size: 3, Total: 4, SkipCount: true}
				p.Reset()
				if p.GetPage() != 0 || p.GetSize() != 0 || p.GetTotal() != 0 || p.GetSkipCount() {
					t.Fatalf("pagination result reset = %+v", p)
				}
			},
			descriptor: (&PaginationResult{}).Descriptor,
			check: func() {
				p := &PaginationResult{Page: 2, Size: 3, Total: 4, SkipCount: true}
				if p.String() == "" || p.ProtoReflect().Descriptor().FullName() != "pagination.PaginationResult" {
					t.Fatal("pagination result proto reflection failed")
				}
				p.ProtoMessage()
				if p.GetPage() != 2 || p.GetSize() != 3 || p.GetTotal() != 4 || !p.GetSkipCount() {
					t.Fatalf("pagination result getters = %+v", p)
				}
			},
			nilCheck: func() {
				var p *PaginationResult
				if p.ProtoReflect().Descriptor().FullName() != "pagination.PaginationResult" || p.GetPage() != 0 || p.GetSize() != 0 || p.GetTotal() != 0 || p.GetSkipCount() {
					t.Fatal("nil pagination result getters/reflection failed")
				}
			},
		},
		{
			name: "scroll",
			msg:  &Scroll{Size: 3, Cursor: "c", Expire: 4},
			reset: func() {
				s := &Scroll{Size: 3, Cursor: "c", Expire: 4}
				s.Reset()
				if s.GetSize() != 0 || s.GetCursor() != "" || s.GetExpire() != 0 {
					t.Fatalf("scroll reset = %+v", s)
				}
			},
			descriptor: (&Scroll{}).Descriptor,
			check: func() {
				s := &Scroll{Size: 3, Cursor: "c", Expire: 4}
				if s.String() == "" || s.ProtoReflect().Descriptor().FullName() != "pagination.Scroll" {
					t.Fatal("scroll proto reflection failed")
				}
				s.ProtoMessage()
				if s.GetSize() != 3 || s.GetCursor() != "c" || s.GetExpire() != 4 {
					t.Fatalf("scroll getters = %+v", s)
				}
			},
			nilCheck: func() {
				var s *Scroll
				if s.ProtoReflect().Descriptor().FullName() != "pagination.Scroll" || s.GetSize() != 0 || s.GetCursor() != "" || s.GetExpire() != 0 {
					t.Fatal("nil scroll getters/reflection failed")
				}
			},
		},
		{
			name: "scroll_result",
			msg:  &ScrollResult{Size: 3, NextCursor: "n", Expire: 4},
			reset: func() {
				s := &ScrollResult{Size: 3, NextCursor: "n", Expire: 4}
				s.Reset()
				if s.GetSize() != 0 || s.GetNextCursor() != "" || s.GetExpire() != 0 {
					t.Fatalf("scroll result reset = %+v", s)
				}
			},
			descriptor: (&ScrollResult{}).Descriptor,
			check: func() {
				s := &ScrollResult{Size: 3, NextCursor: "n", Expire: 4}
				if s.String() == "" || s.ProtoReflect().Descriptor().FullName() != "pagination.ScrollResult" {
					t.Fatal("scroll result proto reflection failed")
				}
				s.ProtoMessage()
				if s.GetSize() != 3 || s.GetNextCursor() != "n" || s.GetExpire() != 4 {
					t.Fatalf("scroll result getters = %+v", s)
				}
			},
			nilCheck: func() {
				var s *ScrollResult
				if s.ProtoReflect().Descriptor().FullName() != "pagination.ScrollResult" || s.GetSize() != 0 || s.GetNextCursor() != "" || s.GetExpire() != 0 {
					t.Fatal("nil scroll result getters/reflection failed")
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.check()
			tc.reset()
			tc.nilCheck()
			raw, idx := tc.descriptor()
			if len(raw) == 0 || len(idx) != 1 {
				t.Fatalf("descriptor = %d, %#v", len(raw), idx)
			}
			if tc.msg.ProtoReflect().Descriptor().FullName() == "" {
				t.Fatal("empty full name")
			}
		})
	}
	if len(file_dbx_pagination_pagination_proto_rawDescGZIP()) == 0 {
		t.Fatal("raw descriptor gzip is empty")
	}
	file_dbx_pagination_pagination_proto_init()
}
