package clause

type SelectType string

const (
	SelectTypeRaw SelectType = "raw"
	SelectTypeAvg SelectType = "avg"
	SelectTypeSum SelectType = "sum"
	SelectTypeMax SelectType = "max"
	SelectTypeMin SelectType = "min"
	SelectTypeCnt SelectType = "count"
)

type Select struct {
	Field      string     `json:"field"`
	SelectType SelectType `json:"select_type"`
}
