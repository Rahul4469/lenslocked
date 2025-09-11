package main

import (
	"fmt"

	"github.com/Rahul4469/lenslocked/models"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func main() {
	cfg := models.DefaultPostgresconfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to database")

	// // Create a table
	// _, err = db.Exec(`
	// CREATE TABLE IF NOT EXISTS users(
	// id SERIAL PRIMARY KEY,
	// email TEXT UNIQUE NOT NULL,
	// password_hash TEXT NOT NULL
	// );
	// `)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Tables created.")

	us := models.UserService{
		DB: db,
	}

	user, err := us.Create("rahul1@email.com", "rahul123")
	if err != nil {
		panic(err)
	}

	fmt.Println(user)

	// // Create a table
	// _, err = db.Exec(`
	// CREATE TABLE IF NOT EXISTS users(
	// id SERIAL PRIMARY KEY,
	// email TEXT UNIQUE NOT NULL,
	//  password_hash TEXT NOT NULL
	// );
	// `)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Tables created.")

	// // Inserting data to PGs tables
	// name := "rahul"
	// email := "rahul2@gmail.com"
	// row := db.QueryRow(`
	// 	INSERT INTO users (name, email)
	// 	VALUES ($1, $2) RETURNING id;`, name, email)
	// var id int
	// err = row.Scan(&id)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("User created. id =", id)

}
