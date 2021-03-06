package metrics

import "time"

// FmtDateKey is used to keep the map DateColMap strings consistant
const FmtDateKey = "2006-01-02"

// ColTotalMap represents a map column names to their integer amounts
type ColTotalMap map[string]int

// DateColMap represents a list of Dates (string formatted with "2006-01-02") and
// their respective colum totals (ColTotalMap)
type DateColMap map[string]ColTotalMap

// DateColumn - returns int value and true if value found with dateKey and columnName
func (dcm DateColMap) DateColumn(date time.Time, columnName string) (int, bool) {
	val, found := dcm[DateKey(date)][columnName]
	return val, found
}

// NewDateColumnMap - returns a dateColMap initialized useing begin and end dates provided
func NewDateColumnMap(beginDate, endDate time.Time) DateColMap {
	current := time.Date(beginDate.Year(), beginDate.Month(), beginDate.Day(), 0, 0, 0, 0, beginDate.Location())
	dateMap := DateColMap{}
	for current.Before(endDate) {
		dateMap[current.Format(FmtDateKey)] = map[string]int{}
		current = current.AddDate(0, 0, 1)
	}
	return dateMap
}

// DateKey - returns string formated using dateKeyFmt from time provided
func DateKey(t time.Time) string {
	return t.Format(FmtDateKey)
}
