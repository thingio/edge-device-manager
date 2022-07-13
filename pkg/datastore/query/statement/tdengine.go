package statement

import "github.com/thingio/edge-device-manager/pkg/datastore/query/clause"

type TDengineBuilder struct {
	builder
}

func (tb *TDengineBuilder) Select(selects []*clause.Select) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) From(from string) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) Where(where *clause.Where) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) GroupBy(groupBy *clause.GroupBy) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) OrderBy(orderBy *clause.OrderBy) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) Raw(raw string) Builder {
	panic("implement me")
}

func (tb *TDengineBuilder) Statement() string {
	return tb.sb.String()
}
