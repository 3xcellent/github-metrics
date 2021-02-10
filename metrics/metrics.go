package metrics

import (
	"fmt"
	"math"
	"time"
)

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
