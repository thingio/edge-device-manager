package clause

type OrderBy struct {
	Field string `json:"field"` // Order by which field
	Desc  bool   `json:"desc"`  // DESCENDING OR ASCENDING
}
