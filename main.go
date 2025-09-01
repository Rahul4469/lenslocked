package main

import (
	"fmt"
	"net/http"

	"github.com/Rahul4469/lenslocked/controllers"
	"github.com/Rahul4469/lenslocked/templates"
	"github.com/Rahul4469/lenslocked/views"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	tpl, err := views.ParseFS(templates.FS, "home.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/", controllers.StaticHandler(tpl))

	tpl, err = views.ParseFS(templates.FS, "contact.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/contact", controllers.StaticHandler(tpl))

	tpl, err = views.ParseFS(templates.FS, "faq.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.FAQ(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	fmt.Println("Starting server at port 3000...")
	http.ListenAndServe(":3000", r)
}
