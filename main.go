package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

type Book struct {
    ID     int    `json:"id,omitempty"`
    Title  string `json:"title"`
    Author string `json:"author"`
    Year   int    `json:"year"`
}


var db *sql.DB

func initDB() {
	var err error
	connStr := "postgres://frendl1:An1metop@localhost:5432/dbweb?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("База данных недоступна:", err)
	}

	log.Println("Подключение к базе данных успешно.")
}

func getBooks(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT id, title, author, year FROM books")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Ошибка выполнения запроса")
	}
	defer rows.Close()

	var books []Book
	for rows.Next() {
		var book Book
		if err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Ошибка при оброботке")
		}
		books = append(books, book)
	}
	return c.JSON(books)
}

func searchBook(c *fiber.Ctx) error {
	id := c.Params("id")
	var book Book

	err := db.QueryRow("SELECT id, title, author, year FROM books WHERE id=$1", id).Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).SendString("Книга не найдена")
	} else if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Некотректный ID")
	}

	return c.JSON(book)
}

func addBook(c *fiber.Ctx) error {
	var book Book
	if err := c.BodyParser(&book); err != nil {
		log.Printf("Парсер тупит: %v\n", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"Некоректные данные, проверьте формат JSON",
		})
	}

	err := db.QueryRow("INSERT INTO books(title, author, year) VALUES ($1, $2, $3) RETURNING id", book.Title, book.Author, book.Year).Scan(&book.ID)
	if err != nil {
		log.Printf("Ошибка при добовлении книги в базу данных: %v\n", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":"Не удалось добавить книгу в базу данных",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(book)

}

func updateBook(c *fiber.Ctx) error{
	id:=c.Params("id")
	var book Book
	if err := c.BodyParser(&book); err!=nil{
		return	c.Status(fiber.StatusBadRequest).SendString("Некоректный запрос")
	}
	_, err := db.Exec("UPDATE books SET title=$1, author=$2, year=$3 WHERE id=$4", book.Title, book.Author, book.Year, id)
	if err!=nil{
		log.Printf("книга не обновлена потому что: %v\n", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString("Книга не обновлена")
	}
	
	return c.SendString("Книга обновлена")
}

func deleteBook(c *fiber.Ctx) error{
	id:= c.Params("id")
	_,err:=db.Exec("DELETE FROM books WHERE id=$1",id)
	if err!= nil{
		return c.Status(fiber.StatusBadRequest).SendString("Некоректный ID")
	}

	return c.SendString("Книга не найдена")

}

func main() {
	initDB()
	defer db.Close()

	app := fiber.New()

	app.Get("/books", getBooks)
	app.Get("/books/:id", searchBook)
	app.Post("/books", addBook)
	app.Put("/books/:id",updateBook)
	app.Delete("/books/:id",deleteBook)

	log.Fatal(app.Listen(":8080"))

}
