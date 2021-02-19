package metrics

import "time"

var ColumnNameMap = map[string]int{}

type ColumnAmount struct {
	Name   string
	Amount int
}

type ColumnAmounts []ColumnAmount

type ColumnsMetric struct {
	Date          time.Time
	ColumnAmounts ColumnAmounts
}
