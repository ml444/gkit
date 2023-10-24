package listoption

func (p *Paginate) SetCurrentPage(page uint32) *Paginate {
	p.Page = page
	if p.Size > 0 {
		p.Offset = (page - 1) * p.Size
	}
	return p
}

func (p *Paginate) SetPageSize(size uint32) *Paginate {
	p.Size = size
	return p
}

func (p *Paginate) SetOffset(offset uint32) *Paginate {
	p.Offset = offset
	return p
}

func (p *Paginate) SetLimit(limit uint32) *Paginate {
	p.Size = limit
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
