package paging

func NewDefaultPaginate() *Paginate {
	return &Paginate{
		Page:      1,
		Size:      10,
		Total:     0,
		SkipCount: false,
	}
}

// SetPage method set current page
func (p *Paginate) SetPage(page uint32) *Paginate {
	p.Page = page
	if page == 0 {
		p.Page = 1
	}
	return p
}

// SetNextPage If it is the last page, return false
func (p *Paginate) SetNextPage() (ok bool) {
	ok = true
	if p.Page == 0 {
		p.Page = 1
	}
	p.Page += 1
	if int64(p.Page)*int64(p.Size) >= p.Total {
		ok = false
	}
	return ok
}

func (p *Paginate) SetSize(size uint32) *Paginate {
	p.Size = size
	return p
}

func (p *Paginate) SetPageAndSize(page, size uint32) *Paginate {
	p.SetPage(page).SetSize(size)
	return p
}

func (p *Paginate) SetSkipCount() *Paginate {
	p.SkipCount = true
	return p
}

func (p *Paginate) Offset() int {
	return int(p.Size * p.Page)
}

func (x *Scroll) SetSize(size uint32) *Scroll {
	x.Size = size
	return x
}

func (x *Scroll) SetToken(token string) *Scroll {
	x.Token = token
	return x
}
