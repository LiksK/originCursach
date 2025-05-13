package main

import (
	"litelife/database" // для работы с бд
	"litelife/handlers" // для работы с обработчиками
	"log"               // для логирования
	"net/http"          // для работы с http
)

func main() {

	database.InitDB()         // инициализация бд из database.go
	defer database.DB.Close() // закрытие бд отложить выполнение функции до момента завершения выполнения верхней функции

	handlers.InitTemplates() // инициализация шаблонов из handlers.go

	fs := http.FileServer(http.Dir("static"))                 // сервер для статических файлов из папки static
	http.Handle("/static/", http.StripPrefix("/static/", fs)) // обработка запросов к статическим файлам

	//Каждому URL сопоставляется функция-обработчик из пакета handlers
	http.HandleFunc("/", handlers.LoginHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginProcessHandler)
	http.HandleFunc("/register-process", handlers.RegisterProcessHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/user", handlers.UserIndexHandler)
	http.HandleFunc("/build-request", handlers.BuildRequestHandler)
	http.HandleFunc("/submit-request", handlers.SubmitRequestHandler)
	http.HandleFunc("/room-booking", handlers.RoomBookingHandler)
	http.HandleFunc("/submit-booking", handlers.SubmitBookingHandler)
	http.HandleFunc("/send-message", handlers.SendMessageHandler)
	http.HandleFunc("/ws", handlers.HandleWebSocket)
	http.HandleFunc("/admin", handlers.AdminIndexHandler)
	http.HandleFunc("/admin-build-requests", handlers.AdminRequestsHandler)
	http.HandleFunc("/admin/approve-request", handlers.ApproveRequestHandler)
	http.HandleFunc("/admin-room-booking", handlers.AdminRoomBookingsHandler)
	http.HandleFunc("/admin/approve-booking", handlers.ApproveBookingHandler)
	http.HandleFunc("/admin/reject-booking", handlers.RejectBookingHandler)
	http.HandleFunc("/admin/revert-request", handlers.RevertRequestHandler)
	http.HandleFunc("/admin/revert-booking", handlers.RevertBookingHandler)
	http.HandleFunc("/admin/delete-message", handlers.DeleteMessageHandler)

	// поток, который отвечает за рассылку сообщений всем подключённым клиентам через WebSocket для чата
	go handlers.BroadcastMessages()

	// запуск сервака на порту 8080
	log.Println("Сервак на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
