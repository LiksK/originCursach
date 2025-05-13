package models
// пакет для работы с моделями
import ( 
	"database/sql"// пакет для работы с базой данных
	"time" // пакет для работы с временем

	"github.com/gorilla/websocket" // пакет для работы с websocket
)

type Client struct { // структура для работы с websocket
	Conn    *websocket.Conn // соединение
	Send    chan []byte // канал для отправки сообщений
	Cleanup func() // функция для очистки
}

func (c *Client) ReadPump() { // функция для чтения сообщений
	defer func() { 
		if c.Cleanup != nil { 
			c.Cleanup()
		}
		c.Conn.Close()
	}()

	for { 
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) WritePump() {
	defer c.Conn.Close()

	for {
		message, ok := <-c.Send
		if !ok {
			c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}

type User struct {
	ID       int
	Username string
	Password string
}

type PageData struct {
	Error         string
	Username      string
	Requests      []RepairRequest
	Messages      []ChatMessage
	Bookings      []RoomBooking
	Stats         []DailyStats
	Notifications []Notification
}

type RepairRequest struct {
	ID         int
	Name       string
	Username   string
	Apartment  string
	RepairType string
	Comment    string
	CreatedAt  time.Time
	IsApproved bool
}

type ChatMessage struct {
	ID        int
	Username  string
	Message   string
	CreatedAt time.Time
}

type RoomBooking struct {
	ID          int
	Name        string
	Phone       string
	RoomNumber  int
	BookingDate time.Time
	IsApproved  sql.NullBool
	CreatedAt   time.Time
}

type DailyStats struct {
	Date         time.Time
	RepairCount  int
	BookingCount int
}

type Notification struct {
	ID        int
	Username  string
	Message   string
	IsRead    bool
	CreatedAt time.Time
}
