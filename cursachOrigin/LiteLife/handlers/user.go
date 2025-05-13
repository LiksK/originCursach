package handlers

import (
	"litelife/database"
	"litelife/models"
	"net/http"
)

func UserIndexHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Load repair requests
	rows, err := database.DB.Query(`
		SELECT id, name, apartment, repair_type, comment, created_at, is_approved
		FROM repair_requests
		WHERE username = $1
		ORDER BY created_at DESC
	`, username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []models.RepairRequest
	for rows.Next() {
		var req models.RepairRequest
		err := rows.Scan(&req.ID, &req.Name, &req.Apartment, &req.RepairType,
			&req.Comment, &req.CreatedAt, &req.IsApproved)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		requests = append(requests, req)
	}

	// Load room bookings
	bookingRows, err := database.DB.Query(`
		SELECT id, name, phone, room_number, booking_date, is_approved, created_at
		FROM room_bookings
		WHERE name = $1
		ORDER BY created_at DESC
	`, username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer bookingRows.Close()

	var bookings []models.RoomBooking
	for bookingRows.Next() {
		var booking models.RoomBooking
		err := bookingRows.Scan(&booking.ID, &booking.Name, &booking.Phone,
			&booking.RoomNumber, &booking.BookingDate, &booking.IsApproved,
			&booking.CreatedAt)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	// Load notifications
	notificationRows, err := database.DB.Query(`
		SELECT id, username, message, is_read, created_at
		FROM notifications
		WHERE username = $1
		ORDER BY created_at DESC
	`, username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer notificationRows.Close()

	var notifications []models.Notification
	for notificationRows.Next() {
		var notification models.Notification
		err := notificationRows.Scan(&notification.ID, &notification.Username,
			&notification.Message, &notification.IsRead, &notification.CreatedAt)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		notifications = append(notifications, notification)
	}

	// Mark notifications as read
	_, err = database.DB.Exec(`
		UPDATE notifications
		SET is_read = true
		WHERE username = $1 AND is_read = false
	`, username)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	messages, err := LoadMessages()
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	data := models.PageData{
		Username:      username,
		Requests:      requests,
		Messages:      messages,
		Bookings:      bookings,
		Notifications: notifications,
	}
	templates.ExecuteTemplate(w, "userindex.html", data)
}
