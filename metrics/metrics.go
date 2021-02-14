package metrics

import (
	"fmt"
	"math"
	"time"
)

var AvailableMetrics = Metrics{
	{Name: "columns", Description: "List of dates with Number of cards in each column for each date"},
	{Name: "issues", Description: "List of issues and their development history and calculated dev and blocked time"},
}

type Metric struct {
	Name        string
	Description string
}

type Metrics []Metric

func FmtDays(d time.Duration) string {
	d = d.Round(time.Minute)
	days := int(math.Ceil(float64(d / time.Hour / 24)))
	//h := d / time.Hour
	//d -= h * time.Hour
	//m := d / time.Minute
	return fmt.Sprintf("%d", days)
}

func FmtDaysHours(d time.Duration) string {
	d = d.Round(time.Minute)
	duration := float64(d) / float64(time.Hour) / 24
	return fmt.Sprintf("%.1f", duration)
}
