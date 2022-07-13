package clause

type Where struct {
	BeginTimeNano uint64 `json:"begin_time_nano"`
	EndTimeNano   uint64 `json:"end_time_nano" ` // select a time range with BeginTimeNano
	Timestamp     uint64 `json:"timestamp"`      // select a time point
	Advanced      string `json:"advanced"`       // Such as where timestamp=1653903530465990845 and device=randnum_test01
}
