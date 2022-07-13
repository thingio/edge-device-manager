package query

import (
	"github.com/thingio/edge-device-manager/pkg/datastore/query/clause"
	"github.com/thingio/edge-device-manager/pkg/datastore/query/statement"
	"github.com/thingio/edge-device-std/config"
)

type Request struct {
	Selects []*clause.Select `json:"selects"`
	From    string           `json:"from"`
	Where   *clause.Where    `json:"where"`
	GroupBy *clause.GroupBy  `json:"group_by"`
	OrderBy *clause.OrderBy  `json:"order_by"`

	Raw string `json:"raw"`
}

func (r *Request) BuildStatement(dsType config.DataStoreType) string {
	builder := statement.NewBuilder(dsType)
	return builder.
		Select(r.Selects).
		From(r.From).
		Where(r.Where).
		GroupBy(r.GroupBy).
		OrderBy(r.OrderBy).
		Raw(r.Raw).
		Statement()
}


type Response struct {
	Rows []map[string]interface{} `json:"results"`
}
