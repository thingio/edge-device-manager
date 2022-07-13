package statement

import (
	"fmt"
	"github.com/thingio/edge-device-manager/pkg/datastore/query/clause"
	"strings"
)

type InfluxDBBuilder struct {
	builder
}

func (ib *InfluxDBBuilder) Select(selects []*clause.Select) Builder {
	fields := make([]string, 0)
	for _, s := range selects {
		switch s.SelectType {
		case clause.SelectTypeRaw:
			fields = append(fields, fmt.Sprintf("%s", s.Field))
		case clause.SelectTypeMax:
			fields = append(fields, fmt.Sprintf("MAX(%s) as %s", s.Field, s.Field))
		case clause.SelectTypeMin:
			fields = append(fields, fmt.Sprintf("MIN(%s) as %s", s.Field, s.Field))
		case clause.SelectTypeAvg:
			fields = append(fields, fmt.Sprintf("MEAN(%s) as %s", s.Field, s.Field))
		case clause.SelectTypeSum:
			fields = append(fields, fmt.Sprintf("SUM(%s) as %s", s.Field, s.Field))
		case clause.SelectTypeCnt:
			fields = append(fields, fmt.Sprintf("COUNT(%s) as count", s.Field))
		}
	}
	if len(fields) == 0 {
		return ib
	}

	ib.sb.WriteString("SELECT " + strings.Join(fields, ", "))
	return ib
}

func (ib *InfluxDBBuilder) From(from string) Builder {
	if from == "" {
		return ib
	}

	ib.sb.WriteString(" FROM " + from)
	return ib
}

func (ib *InfluxDBBuilder) Where(where *clause.Where) Builder {
	if where == nil {
		return ib
	}

	conditions := make([]string, 0)
	if where.Timestamp != 0 {
		conditions = append(conditions, fmt.Sprintf("time == %d", where.Timestamp))
	}
	if where.BeginTimeNano != 0 {
		conditions = append(conditions, fmt.Sprintf("time >= %d", where.BeginTimeNano))
	}
	if where.EndTimeNano != 0 {
		conditions = append(conditions, fmt.Sprintf("time <= %d", where.EndTimeNano))
	}
	if where.Advanced != "" {
		conditions = append(conditions, where.Advanced)
	}
	if len(conditions) == 0 {
		return ib
	}

	ib.sb.WriteString(" WHERE " + strings.Join(conditions, " AND "))
	return ib
}

func (ib *InfluxDBBuilder) GroupBy(groupBy *clause.GroupBy) Builder {
	if groupBy == nil {
		return ib
	}

	// https://docs.influxdata.com/influxdb/v1.6/query_language/data_exploration/#the-group-by-clause
	sb := strings.Builder{}

	if groupBy.TimeInterval != nil {
		interval := groupBy.TimeInterval.Interval
		offset := groupBy.TimeInterval.OffsetInterval
		if interval != "" {
			if offset == "" {
				sb.WriteString(fmt.Sprintf("time(%s)", interval))
			} else {
				sb.WriteString(fmt.Sprintf("time(%s, %s)", interval, offset))
			}
		}
	}

	if groupBy.Tags != nil {
		if sb.Len() != 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strings.Join(groupBy.Tags, ","))
		sb.WriteByte(' ')
	}

	if groupBy.Fill != nil {
		sb.WriteByte(' ')
		switch groupBy.Fill.FillType {
		case clause.FillTypeNull:
			sb.WriteString(fmt.Sprintf("fill(null)"))
		case clause.FillTypePrev:
			sb.WriteString(fmt.Sprintf("fill(previous)"))
		case clause.FillTypeLinear:
			sb.WriteString(fmt.Sprintf("fill(linear)"))
		case clause.FillTypeNone:
			sb.WriteString(fmt.Sprintf("fill(none)"))
		case clause.FillTypeValue:
			sb.WriteString(fmt.Sprintf("fill(%v)", groupBy.Fill.Value))
		default: // None
		}
	}
	if sb.Len() == 0 {
		return ib
	}

	ib.sb.WriteString(" GROUP BY " + sb.String())
	return ib
}

func (ib *InfluxDBBuilder) OrderBy(orderBy *clause.OrderBy) Builder {
	if orderBy == nil {
		return ib
	}

	// https://docs.influxdata.com/influxdb/v1.6/query_language/data_exploration/#order-by-time-desc
	sb := strings.Builder{}
	if orderBy.Field != "" {
		if orderBy.Field != "time" {
			return ib
		} else {
			sb.WriteString(fmt.Sprintf(" ORDER BY time"))
		}

		if orderBy.Desc {
			sb.WriteString(" DESC")
		} else {
			sb.WriteString(" ASC")
		}
	}

	ib.sb.WriteString(sb.String())
	return ib
}

func (ib *InfluxDBBuilder) Raw(raw string) Builder {
	if raw == "" {
		return ib
	}

	ib.sb.WriteString(raw)
	return ib
}

func (ib *InfluxDBBuilder) Statement() string {
	return ib.sb.String()
}
