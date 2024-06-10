package pagination

func NewDefaultPagination() *Pagination {
	return &Pagination{
		Page:      1,
		Size:      10,
		Total:     0,
		SkipCount: false,
	}
}

// SetPage method set current page
func (x *Pagination) SetPage(page uint32) *Pagination {
	x.Page = page
	if page == 0 {
		x.Page = 1
	}
	return x
}

// SetNextPage If it is the last page, return false
func (x *Pagination) SetNextPage() (ok bool) {
	ok = true
	if x.Page == 0 {
		x.Page = 1
	}
	x.Page += 1
	if int64(x.Page)*int64(x.Size) >= x.Total {
		ok = false
	}
	return ok
}

func (x *Pagination) SetSize(size uint32) *Pagination {
	x.Size = size
	return x
}

func (x *Pagination) SetPageAndSize(page, size uint32) *Pagination {
	x.SetPage(page).SetSize(size)
	return x
}

func (x *Pagination) SetSkipCount() *Pagination {
	x.SkipCount = true
	return x
}

func (x *Pagination) Offset() int {
	if x.Page <= 1 {
		return 0
	}
	return int(x.Size * (x.Page - 1))
}

func (x *Scroll) SetSize(size uint32) *Scroll {
	x.Size = size
	return x
}

func (x *Scroll) SetToken(token string) *Scroll {
	x.Token = token
	return x
}
