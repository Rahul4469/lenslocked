package main

import (
	"fmt"
	"net/http"

	"github.com/Rahul4469/lenslocked/controllers"
	"github.com/Rahul4469/lenslocked/migrations"
	"github.com/Rahul4469/lenslocked/models"
	"github.com/Rahul4469/lenslocked/templates"
	"github.com/Rahul4469/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
)

func main() {
	// Setup the Database ---------------
	cfg := models.DefaultPostgresconfig()
	fmt.Println(cfg.String())
	db, err := models.Open(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup Services ---------------
	userService := models.UserService{
		DB: db,
	}
	sessionService := models.SessionService{
		DB: db,
	}

	// Setup Middleware ---------------
	umw := controllers.UserMiddleware{
		SessionService: &sessionService,
	}

	csrfKey := "Z3xhNej1AqaKKpM4Qx1yGZconAT2NVE0"
	csrfMw := csrf.Protect([]byte(csrfKey), csrf.Secure(false), csrf.Path("/"), csrf.TrustedOrigins([]string{"http://localhost:3000",
		"http://127.0.0.1:3000",
		"localhost:3000",
		"127.0.0.1:3000",
	}))

	// Setup Contollers ---------------
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
	userC.Templates.ForgotPassword, err = views.ParseFS(templates.FS, "forgot-pw.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}

	//---------------------------------------------------
	// Setup Router and Routes ---------------
	r := chi.NewRouter()
	r.Use(csrfMw)
	r.Use(umw.SetUser)

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

	tpl, err = views.ParseFS(templates.FS, "faq.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	r.Get("/faq", controllers.FAQ(tpl))

	r.Get("/signup", userC.New)
	r.Post("/users", userC.Create)
	r.Get("/signin", userC.SignIn)
	r.Post("/signin", userC.ProcessSignIn)
	r.Post("/signout", userC.ProcessSignOut)
	r.Get("/forgot-pw", userC.ForgotPassword)
	r.Post("/forgot-pw", userC.ProcessForgotPassword)
	// r.Get("/users/me", userC.CurrentUser)

	//The r in the callback is a newly created subrouter, scoped to /user/me
	//Chi creates a fresh subrouter—a new, independent routing context
	//All sub routes under this Route() will have the user context data
	//even "/" request after this Route will spawn with the user ctx data
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", userC.CurrentUser)

	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	//-----------------------------------------------------
	// Start the Server
	fmt.Println("Starting server at port 3000...")
	http.ListenAndServe(":3000", r)
	//once starting the router, all these methods are registered on the router
	//one receiving any request these methods are matched and the methods in the arguments gets executed for the response
}
