package paging

func NewDefaultPaginate() *Paginate {
	return &Paginate{
		Page:      1,
		Size:      10,
		Offset:    0,
		Total:     0,
		SkipCount: false,
	}
}

func (p *Paginate) SetCurrentPage(page uint32) *Paginate {
	if page == 0 {
		page = 1
	}
	p.Page = page
	if p.Size > 0 {
		p.Offset = (page - 1) * p.Size
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
	p.Offset = (p.Page - 1) * p.Size
	if int64(p.Page)*int64(p.Size) >= p.Total {
		ok = false
	}
	return ok
}

func (p *Paginate) SetPageSize(size uint32) *Paginate {
	p.Size = size
	if p.Page > 1 && p.Offset == 0 {
		p.Offset = (p.Page - 1) * p.Size
	}
	return p
}

func (p *Paginate) SetOffset(offset uint32) *Paginate {
	p.Offset = offset
	if p.Size > 0 {
		p.Page = (p.Offset / p.Size) + 1
		// NOTE: when offset and page are mixed, duplicate data is more acceptable than missing data.
		//if p.Offset % p.Size > 0 {
		//	p.Page += 1
		//}
	}
	return p
}

func (p *Paginate) SetLimit(limit uint32) *Paginate {
	p.Size = limit
	if p.Page > 1 && p.Offset == 0 {
		p.Offset = (p.Page - 1) * p.Size
	}
	return p
}

func (p *Paginate) SetOffsetLimit(offset, limit uint32) *Paginate {
	p.SetLimit(limit).SetOffset(offset)
	return p
}
func (p *Paginate) SetSkipCount() *Paginate {
	p.SkipCount = true
	return p
}

func (x *Scroll) SetSize(size uint32) *Scroll {
	x.Size = size
	return x
}
func (x *Scroll) SetNextToken(token string) *Scroll {
	x.NextToken = token
	return x
}
