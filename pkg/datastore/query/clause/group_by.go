package clause

type GroupBy struct {
	Fill         *Fill         `json:"fill"`
	TimeInterval *TimeInterval `json:"time_interval"`
	Tags         []string      `json:"tags"`
}

func (g *GroupBy) AddTag(tag string) *GroupBy {
	g.Tags = append(g.Tags, tag)
	return g
}

func (g *GroupBy) SetFill(fill *Fill) *GroupBy {
	g.Fill = fill
	return g
}

func (g *GroupBy) SetTimeInterval(interval *TimeInterval) *GroupBy {
	g.TimeInterval = interval
	return g
}
