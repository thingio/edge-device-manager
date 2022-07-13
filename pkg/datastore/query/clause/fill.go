package clause

type FillType string

const (
	FillTypeNone   FillType = "none"   // 不返回没有值的数据
	FillTypePrev   FillType = "prev"   // 填充前一个区间的数据
	FillTypeNext   FillType = "next"   // 填充后一个区间的数据
	FillTypeLinear FillType = "linear" //
	FillTypeNull   FillType = "null"   // 填充空值
	FillTypeValue  FillType = "value"  // 填充指定的值
)

type Fill struct {
	Value    float64  `json:"value"`
	FillType FillType `json:"fill_type"`
}

func (f *Fill) SetValue(value float64) *Fill {
	f.Value = value
	return f
}
