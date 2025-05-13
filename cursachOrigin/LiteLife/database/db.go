package database
//созхдание пакета бд

import ( //иморты 
	"database/sql"// пакет для работы с бд
	"log"// пакет для ощиболк

	_ "github.com/lib/pq"//  драйвер для работы с бд
)

var DB *sql.DB// переменная для работы с бд

func InitDB() { // функция для инициализации бд
	log.Println("Initializing database connection...")
	connStr := "user=postgres dbname=litelifedb password=Pgadmin port=5432 sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database connection: %v", err)
	}

	// Проверяем соединение
	err = DB.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	log.Println("Database connection established successfully")

	createTables()
}

func createTables() {
	log.Println("Creating database tables...")

	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL
		)
	`)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}
	log.Println("Users table created/verified")

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS repair_requests (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			username VARCHAR(50) NOT NULL,
			apartment VARCHAR(20) NOT NULL,
			repair_type VARCHAR(50) NOT NULL,
			comment TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_approved BOOLEAN DEFAULT FALSE
		)
	`)
	if err != nil {
		log.Fatalf("Error creating repair_requests table: %v", err)
	}
	log.Println("Repair requests table created/verified")

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS chat_messages (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			message TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating chat_messages table: %v", err)
	}
	log.Println("Chat messages table created/verified")

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS room_bookings (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			phone VARCHAR(20) NOT NULL,
			room_number INTEGER NOT NULL,
			booking_date DATE NOT NULL,
			is_approved BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating room_bookings table: %v", err)
	}
	log.Println("Room bookings table created/verified")

	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			message TEXT NOT NULL,
			is_read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatalf("Error creating notifications table: %v", err)
	}
	log.Println("Notifications table created/verified")
	log.Println("All database tables created/verified successfully")
}
