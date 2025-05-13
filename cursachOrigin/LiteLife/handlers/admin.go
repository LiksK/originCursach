package handlers

import (
	"litelife/database"
	"litelife/models"
	"log"
	"net/http"
	"sort"
	"time"
)

func getLast5DaysStats() ([]models.DailyStats, error) {
	var stats []models.DailyStats
	
	// Get current date and calculate date 5 days ago
	now := time.Now()
	fiveDaysAgo := now.AddDate(0, 0, -4) // -4 because we want to include today
	
	// Create a map to store stats for each day
	statsMap := make(map[string]*models.DailyStats)
	
	// Initialize the map with all days
	for d := fiveDaysAgo; !d.After(now); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		statsMap[dateStr] = &models.DailyStats{
			Date:         d,
			RepairCount:  0,
			BookingCount: 0,
		}
	}
	
	// Get repair requests count for each day
	repairRows, err := database.DB.Query(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM repair_requests
		WHERE created_at >= $1
		GROUP BY DATE(created_at)
	`, fiveDaysAgo)
	if err != nil {
		return nil, err
	}
	defer repairRows.Close()
	
	for repairRows.Next() {
		var date time.Time
		var count int
		if err := repairRows.Scan(&date, &count); err != nil {
			return nil, err
		}
		dateStr := date.Format("2006-01-02")
		if stat, exists := statsMap[dateStr]; exists {
			stat.RepairCount = count
		}
	}
	
	// Get booking requests count for each day
	bookingRows, err := database.DB.Query(`
		SELECT DATE(created_at) as date, COUNT(*) as count
		FROM room_bookings
		WHERE created_at >= $1
		GROUP BY DATE(created_at)
	`, fiveDaysAgo)
	if err != nil {
		return nil, err
	}
	defer bookingRows.Close()
	
	for bookingRows.Next() {
		var date time.Time
		var count int
		if err := bookingRows.Scan(&date, &count); err != nil {
			return nil, err
		}
		dateStr := date.Format("2006-01-02")
		if stat, exists := statsMap[dateStr]; exists {
			stat.BookingCount = count
		}
	}
	
	// Convert map to slice and sort by date
	for _, stat := range statsMap {
		stats = append(stats, *stat)
	}
	
	// Sort stats by date
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Date.Before(stats[j].Date)
	})
	
	return stats, nil
}

func AdminIndexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok || username != "admin" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get statistics for the last 5 days
	stats, err := getLast5DaysStats()
	if err != nil {
		log.Printf("Error getting statistics: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Println("Fetching repair requests...")
	rows, err := database.DB.Query(`
		SELECT id, name, username, apartment, repair_type, comment, created_at, is_approved
		FROM repair_requests
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Error fetching repair requests: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []models.RepairRequest
	for rows.Next() {
		var req models.RepairRequest
		err := rows.Scan(&req.ID, &req.Name, &req.Username, &req.Apartment,
			&req.RepairType, &req.Comment, &req.CreatedAt, &req.IsApproved)
		if err != nil {
			log.Printf("Error scanning repair request: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		requests = append(requests, req)
	}

	log.Println("Loading messages...")
	messages, err := LoadMessages()
	if err != nil {
		log.Printf("Error loading messages: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	log.Println("Fetching room bookings...")
	bookingsRows, err := database.DB.Query(`
		SELECT id, name, phone, room_number, booking_date, is_approved, created_at
		FROM room_bookings
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Printf("Error fetching room bookings: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer bookingsRows.Close()

	var bookings []models.RoomBooking
	for bookingsRows.Next() {
		var booking models.RoomBooking
		err := bookingsRows.Scan(&booking.ID, &booking.Name, &booking.Phone,
			&booking.RoomNumber, &booking.BookingDate, &booking.IsApproved,
			&booking.CreatedAt)
		if err != nil {
			log.Printf("Error scanning room booking: %v", err)
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	log.Println("Rendering admin template...")
	data := models.PageData{
		Username: username,
		Requests: requests,
		Messages: messages,
		Bookings: bookings,
		Stats:    stats,
	}
	err = templates.ExecuteTemplate(w, "adminindex.html", data)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	log.Println("Admin page rendered successfully")
}
