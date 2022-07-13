package statement

import (
	"github.com/thingio/edge-device-manager/pkg/datastore/query/clause"
	"github.com/thingio/edge-device-std/config"
	"strings"
)

func NewBuilder(dsType config.DataStoreType) Builder {
	super := builder{
		sb: strings.Builder{},
	}

	var b Builder
	switch dsType {
	case config.DataStoreTypeInfluxDB:
		ib := &InfluxDBBuilder{super}
		b = ib
	case config.DataStoreTypeTDengine:
		tb := &TDengineBuilder{super}
		b = tb
	}

	return b
}

type Builder interface {
	Select(selects []*clause.Select) Builder
	From(from string) Builder
	Where(where *clause.Where) Builder
	GroupBy(groupBy *clause.GroupBy) Builder
	OrderBy(orderBy *clause.OrderBy) Builder
	Raw(raw string) Builder

	Statement() string
}

type builder struct {
	sb strings.Builder
}
