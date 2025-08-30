package main

import (
	"fmt"
	"html/template"
	"os"
)

type User struct {
	Name string
	Age  int
	Bio  string
}

func main() {
	tpl, err := template.ParseFiles("user.gohtml")
	if err != nil {
		panic(err)
	}

	user := User{
		Name: "Rahul",
		Age:  19,
		Bio:  "unemployed",
	}

	err = tpl.Execute(os.Stdout, user)
	if err != nil {
		panic(err)
	}
	fmt.Println(err)
}
