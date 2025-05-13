package handlers
// пакет для обработки запросов на ремонт
import (
	"net/http"// пакет для работы с http
	"litelife/models"// пакет для работы с моделями
	"litelife/database"// пакет для работы с базой данных
)

func BuildRequestHandler(w http.ResponseWriter, r *http.Request) { // функция для обработки запросов на ремонт
	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на главную страницу
		return
	}

	data := models.PageData{Username: username} // инициализация data
	templates.ExecuteTemplate(w, "buildRequest.html", data) // выполнение шаблона
}

func SubmitRequestHandler(w http.ResponseWriter, r *http.Request) { // функция для отправки запроса на ремонт
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/build-request", http.StatusSeeOther) // перенаправление на страницу запроса на ремонт
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на главную страницу
		return
	}

	name := r.FormValue("name") // получение имени пользователя из формы
	apartment := r.FormValue("apartment") // получение номера квартиры из формы
	repairType := r.FormValue("repairType") // получение типа ремонта из формы
	comment := r.FormValue("comment") // получение комментария из формы

	_, err := database.DB.Exec(`
		INSERT INTO repair_requests (name, username, apartment, repair_type, comment)
		VALUES ($1, $2, $3, $4, $5)
	`, name, username, apartment, repairType, comment) // выполнение запроса на ремонт

	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	http.Redirect(w, r, "/user", http.StatusSeeOther) 
}

func AdminRequestsHandler(w http.ResponseWriter, r *http.Request) { // функция для обработки запросов на ремонт
	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	rows, err := database.DB.Query(`
		SELECT id, name, username, apartment, repair_type, comment, created_at, is_approved
		FROM repair_requests
		ORDER BY created_at DESC
	`)
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}
	defer rows.Close() // закрытие rows

	var requests []models.RepairRequest
	for rows.Next() {
		var req models.RepairRequest
		err := rows.Scan(&req.ID, &req.Name, &req.Username, &req.Apartment, 
			&req.RepairType, &req.Comment, &req.CreatedAt, &req.IsApproved) // сканирование строк
		if err != nil { // проверка на ошибку
			http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
			return
		}
		requests = append(requests, req) // добавление запроса в requests
	}

	data := models.PageData{ // инициализация data
		Username: username, // инициализация имени пользователя
		Requests: requests, // инициализация requests
	}
	templates.ExecuteTemplate(w, "adminBuildRequests.html", data) // выполнение шаблона
}

func ApproveRequestHandler(w http.ResponseWriter, r *http.Request) { // функция для одобрения запроса на ремонт
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/admin/requests", http.StatusSeeOther) // перенаправление на страницу запросов на ремонт
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на главную страницу
		return
	}

	requestID := r.FormValue("request_id") // получение id запроса
	
	var requestUsername string // инициализация requestUsername
	err := database.DB.QueryRow("SELECT username FROM repair_requests WHERE id = $1", requestID).Scan(&requestUsername) // получение имени пользователя из запроса
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	_, err = database.DB.Exec("UPDATE repair_requests SET is_approved = true WHERE id = $1", requestID) // обновление статуса запроса
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	_, err = database.DB.Exec(` 
		INSERT INTO notifications (username, message)
		VALUES ($1, $2)
	`, requestUsername, "Ваша заявка на ремонт была одобрена администратором") // создание уведомления

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/requests", http.StatusSeeOther)
}

func RevertRequestHandler(w http.ResponseWriter, r *http.Request) { // функция для отмены запроса на ремонт	
	if r.Method != http.MethodPost { // проверка на метод запроса
		http.Redirect(w, r, "/admin/requests", http.StatusSeeOther) // перенаправление на страницу запросов на ремонт
		return
	}

	session, _ := store.Get(r, "session-name") // получение сессии
	username, ok := session.Values["username"].(string) // получение имени пользователя из сессии
	if !ok || username != "admin" { // проверка на ошибку
		http.Redirect(w, r, "/", http.StatusSeeOther) // перенаправление на главную страницу
		return
	}

	requestID := r.FormValue("request_id") // получение id запроса
	_, err := database.DB.Exec("UPDATE repair_requests SET is_approved = false WHERE id = $1", requestID) // обновление статуса запроса
	if err != nil { // проверка на ошибку
		http.Error(w, "Server error", http.StatusInternalServerError) // вывод ошибки
		return
	}

	http.Redirect(w, r, "/admin/requests", http.StatusSeeOther)
} 