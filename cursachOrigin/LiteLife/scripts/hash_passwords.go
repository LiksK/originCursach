package main

import (
	"database/sql"    // Пакет для работы с SQL базами данных
	"fmt"            // Пакет для форматированного вывода
	"log"            // Пакет для логирования
	"golang.org/x/crypto/bcrypt" // Пакет для хэширования паролей
	_ "github.com/lib/pq"        // Драйвер PostgreSQL
)

func main() {
	// Строка подключения к базе данных PostgreSQL
	// Указываем параметры: пользователь, имя базы данных, пароль, порт и отключение SSL
	connStr := "user=postgres dbname=litelifedb password=Pgadmin port=5432 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close() // Закрываем соединение с базой данных при завершении программы

	// Выполняем SQL-запрос для получения всех пользователей из таблицы users
	// Выбираем id, имя пользователя и пароль
	rows, err := db.Query("SELECT id, username, password FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close() // Закрываем курсор после завершения работы

	// Перебираем всех пользователей из результата запроса
	for rows.Next() {
		var id int
		var username, password string
		// Сканируем значения из текущей строки в переменные
		err := rows.Scan(&id, &username, &password)
		if err != nil {
			log.Printf("Ошибка при чтении данных пользователя: %v", err)
			continue
		}

		// Генерируем хэш пароля с использованием bcrypt
		// bcrypt.DefaultCost определяет сложность хэширования
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Ошибка при хэшировании пароля для пользователя %s: %v", username, err)
			continue
		}

		// Обновляем хэшированный пароль в базе данных
		// Используем параметризованный запрос для безопасности
		_, err = db.Exec("UPDATE users SET password = $1 WHERE id = $2", string(hashedPassword), id)
		if err != nil {
			log.Printf("Ошибка при обновлении пароля для пользователя %s: %v", username, err)
			continue
		}

		// Выводим сообщение об успешном обновлении пароля
		fmt.Printf("Обновлен пароль для пользователя: %s\n", username)
	}

	// Выводим сообщение о завершении процесса хэширования
	fmt.Println("Хэширование паролей завершено")
} 