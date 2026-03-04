package cron

import (
	"log"
	"schedulehelper/db"
	"time"
)

func StartCronJobs() {
	go func() {
		for {
			cleanupOldData()
			time.Sleep(7 * 24 * time.Hour)
		}
	}()
}

func cleanupOldData() {
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	log.Printf("Running cleanup for data older than %v", sixMonthsAgo)

	res, err := db.DB.Exec("DELETE FROM slots WHERE end_time < ?", sixMonthsAgo)
	if err == nil {
		count, _ := res.RowsAffected()
		log.Printf("Cleaned up %d old slots", count)
	}

	res, err = db.DB.Exec("DELETE FROM events WHERE end_time < ?", sixMonthsAgo)
	if err == nil {
		count, _ := res.RowsAffected()
		log.Printf("Cleaned up %d old events", count)
	}
}