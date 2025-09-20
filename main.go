package main

import (
	"fmt"
	"net/http"

	"github.com/Rahul4469/lenslocked/controllers"
	"github.com/Rahul4469/lenslocked/models"
	"github.com/Rahul4469/lenslocked/templates"
	"github.com/Rahul4469/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {
	r := chi.NewRouter()
	tpl, err := views.ParseFS(templates.FS, "home.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/", controllers.StaticHandler(tpl))

	tpl, err = views.ParseFS(templates.FS, "contact.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/contact", controllers.StaticHandler(tpl))

	//Create and save new user through signup
	//Open new DB connection -> pass a reference to User controller
	cfg := models.DefaultPostgresconfig()
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	userC := controllers.Users{
		UserService:    &userService, //passing an address
		SessionService: &sessionService,
	}
	userC.Templates.New, err = views.ParseFS(templates.FS, "signup.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	userC.Templates.SignIn, err = views.ParseFS(templates.FS, "signin.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/signup", userC.New)
	r.Post("/users", userC.Create)
	r.Get("/signin", userC.SignIn)
	r.Post("/signin", userC.ProcessSignIn)
	r.Get("/users/me", userC.CurrentUser)

	tpl, err = views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.FAQ(tpl))

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	csrfKey := "Z3xhNej1AqaKKpM4Qx1yGZconAT2NVE0"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false), csrf.Path("/"), csrf.TrustedOrigins([]string{"http://localhost:3000",
		"http://127.0.0.1:3000",
		"localhost:3000",
		"127.0.0.1:3000",
	}))

	fmt.Println("Starting server at port 3000...")
	http.ListenAndServe(":3000", csrfMw(r))
	//once starting the router, all these methods are registered on the router
	//one receiving any request these methods are matched and the methods in the arguments gets executed for the response
}
