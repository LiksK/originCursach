package handlers
// пакет для обработки запросов на бронирование
import (
	"litelife/database"// пакет для работы с бд
	"litelife/models"// пакет для работы с моделями
	"net/http"// пакет для работы с http
)

func RoomBookingHandler(w http.ResponseWriter, r *http.Request) { // функция для обработки запросов на бронирование
	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на страницу авторизации
		return
	}

	data := models.PageData{Username: username} // создание структуры для передачи данных в шаблон
	templates.ExecuteTemplate(w, "roomBooking.html", data) // выполнение шаблона
}

func SubmitBookingHandler(w http.ResponseWriter, r *http.Request) { // функция для обработки запросов на бронирование
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/room-booking", http.StatusSeeOther) // перенаправление на страницу бронирования
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии																				
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name") // получение имени пользователя из формы
	phone := r.FormValue("phone") // получение номера телефона из формы
	roomNumber := r.FormValue("room") // получение номера комнаты из формы
	bookingDate := r.FormValue("booking_date") // получение даты бронирования из формы

	var count int // переменная для хранения количества бронирований
	err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM room_bookings
		WHERE room_number = $1 AND booking_date = $2 AND is_approved = true
	`, roomNumber, bookingDate).Scan(&count) // получение количества бронирований из бд

	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	if count > 0 {
		data := models.PageData{
			Username: username,
			Error:    "Room is already booked for this date",
		}
		templates.ExecuteTemplate(w, "roomBooking.html", data) // выполнение шаблона
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO room_bookings (name, phone, room_number, booking_date)
		VALUES ($1, $2, $3, $4)
	`, name, phone, roomNumber, bookingDate) // вставка бронирования в бд

	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	http.Redirect(w, r, "/user", http.StatusSeeOther) // перенаправление на страницу пользователя
}

func AdminRoomBookingsHandler(w http.ResponseWriter, r *http.Request) { // функция для обработки запросов на бронирование
	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на страницу авторизации
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, name, phone, room_number, booking_date, is_approved, created_at
		FROM room_bookings
		ORDER BY created_at DESC
	`)
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}
	defer rows.Close() // закрытие соединения с бд

	var bookings []models.RoomBooking // переменная для хранения бронирований
	for rows.Next() { // цикл для получения бронирований
		var booking models.RoomBooking // переменная для хранения бронирования
		err := rows.Scan(&booking.ID, &booking.Name, &booking.Phone,
			&booking.RoomNumber, &booking.BookingDate, &booking.IsApproved,
			&booking.CreatedAt)
		if err != nil { // проверка на ошибку
			http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
			return
		}
		bookings = append(bookings, booking)
	}

	data := models.PageData{ // создание структуры для передачи данных в шаблон
		Username: username, // передача имени пользователя в шаблон
		Bookings: bookings, // передача бронирований в шаблон
	}
	templates.ExecuteTemplate(w, "adminRoomBooking.html", data)
}

func ApproveBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // проверка на метод запроса	
		http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther) // перенаправление на страницу бронирований
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на страницу авторизации
		return
	}

	bookingID := r.FormValue("booking_id") // получение id бронирования из формы
	 
	// Get the name of the booking owner
	var bookingName string // переменная для хранения имени бронирования
	err := database.DB.QueryRow("SELECT name FROM room_bookings WHERE id = $1", bookingID).Scan(&bookingName) // получение имени бронирования из бд
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Update the booking status
	_, err = database.DB.Exec("UPDATE room_bookings SET is_approved = true WHERE id = $1", bookingID) // обновление статуса бронирования в бд
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create notification
	_, err = database.DB.Exec(`
		INSERT INTO notifications (username, message)
		VALUES ($1, $2)
	`, bookingName, "Ваша заявка на бронирование помещения была одобрена администратором")

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError) 	
		return
	}

	http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther)
}

func RejectBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther) // перенаправление на страницу бронирований
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на страницу авторизации
		return
	}

	bookingID := r.FormValue("booking_id") // получение id бронирования из формы
	_, err := database.DB.Exec("DELETE FROM room_bookings WHERE id = $1", bookingID) // удаление бронирования из бд
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther) // перенаправление на страницу бронирований
}

func RevertBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther) // перенаправление на страницу бронирований
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на страницу авторизации
		return
	}

	bookingID := r.FormValue("booking_id") // получение id бронирования из формы
	_, err := database.DB.Exec("UPDATE room_bookings SET is_approved = NULL WHERE id = $1", bookingID) // обновление статуса бронирования в бд
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	http.Redirect(w, r, "/admin/bookings", http.StatusSeeOther)
}
