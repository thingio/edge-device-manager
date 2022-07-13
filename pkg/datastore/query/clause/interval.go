package clause

// TimeInterval
// https://docs.taosdata.com/taos-sql/interval/
// https://docs.influxdata.com/influxdb/v1.6/query_language/data_exploration/#group-by-time-intervals
type TimeInterval struct {
	Interval       string `json:"interval"`
	OffsetInterval string `json:"offset_interval"`
	Sliding        string `json:"sliding"`
}
