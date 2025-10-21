package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Rahul4469/lenslocked/controllers"
	"github.com/Rahul4469/lenslocked/migrations"
	"github.com/Rahul4469/lenslocked/models"
	"github.com/Rahul4469/lenslocked/templates"
	"github.com/Rahul4469/lenslocked/views"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/joho/godotenv"
)

type config struct {
	PSQL models.PostgresConfig
	SMTP models.SMTPConfig
	CSRF struct {
		Key            string
		Secure         bool
		TrustedOrigins []string
	}
	Server struct {
		Address string
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}

	cfg.PSQL = models.PostgresConfig{
		Host:     os.Getenv("PSQL_HOST"),
		Port:     os.Getenv("PSQL_PORT"),
		User:     os.Getenv("PSQL_USER"),
		Password: os.Getenv("PSQL_PASSWORD"),
		Database: os.Getenv("PSQL_DATABASE"),
		SSLMode:  os.Getenv("PSQL_SSLMODE"),
	}
	if cfg.PSQL.Host == "" && cfg.PSQL.Port == "" {
		return cfg, fmt.Errorf("no psql config provided")
	}

	cfg.SMTP.Host = os.Getenv("SMTP_HOST")
	portStr := os.Getenv("SMTP_PORT")
	cfg.SMTP.Port, err = strconv.Atoi(portStr)
	if err != nil {
		return cfg, err
	}
	cfg.SMTP.Username = os.Getenv("SMTP_USERNAME")
	cfg.SMTP.Password = os.Getenv("SMTP_PASSWORD")

	cfg.CSRF.Key = os.Getenv("CSRF_KEY")
	cfg.CSRF.Secure = os.Getenv("CSRF_SECURE") == "true"
	cfg.CSRF.TrustedOrigins = strings.Fields(os.Getenv("CSRF_TRUSTED_ORIGINS"))

	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")

	return cfg, nil
}

func main() {
	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	// Setup the Database ---------------
	fmt.Println(cfg.PSQL.String())
	db, err := models.Open(cfg.PSQL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = models.MigrateFS(db, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Setup Services ---------------
	userService := &models.UserService{
		DB: db,
	}
	sessionService := &models.SessionService{
		DB: db,
	}
	pwResetService := &models.PasswordResetService{
		DB: db,
	}
	emailService, err := models.NewEmailService(cfg.SMTP)
	if err != nil {
		panic(err)
	}
	galleryService := &models.GalleryService{
		DB: db,
	}

	// Setup Middleware ---------------
	umw := controllers.UserMiddleware{
		SessionService: sessionService,
	}

	csrfMw := csrf.Protect([]byte(cfg.CSRF.Key), csrf.Secure(cfg.CSRF.Secure), csrf.Path("/"), csrf.TrustedOrigins(cfg.CSRF.TrustedOrigins))

	// Setup Contollers ---------------
	userC := controllers.Users{
		UserService:          userService, //passing an address
		SessionService:       sessionService,
		PasswordResetService: pwResetService,
		EmailService:         emailService,
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
	userC.Templates.CheckYourEmail, err = views.ParseFS(templates.FS, "check-your-email.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	userC.Templates.ResetPassword, err = views.ParseFS(templates.FS, "reset-pw.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	galleriesC := controllers.Galleries{
		GalleryService: galleryService,
	}
	galleriesC.Template.New, err = views.ParseFS(templates.FS, "galleries/new.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	galleriesC.Template.Edit, err = views.ParseFS(templates.FS, "galleries/edit.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	galleriesC.Template.Index, err = views.ParseFS(templates.FS, "galleries/index.gohtml", "tailwind.gohtml")
	if err != nil {
		panic(err)
	}
	galleriesC.Template.Show, err = views.ParseFS(templates.FS, "galleries/show.gohtml", "tailwind.gohtml")
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
	r.Post("/forgot-pw", userC.ProcessForgotPassword) //On button on forgot password page, Check_your_email template will be Executed and rendered
	r.Get("/reset-pw", userC.ResetPassword)
	r.Post("/reset-pw", userC.ProcessResetPassword)
	// r.Get("/users/me", userC.CurrentUser)

	//The r in the callback is a newly created subrouter, scoped to /users/me
	//Chi creates a fresh subrouter—a new, independent routing context
	//All sub routes under this Route() will have the user context data
	//even "/" request after this Route will spawn with the user ctx data
	r.Route("/users/me", func(r chi.Router) {
		r.Use(umw.RequireUser)
		r.Get("/", userC.CurrentUser)

	})
	r.Route("/galleries", func(r chi.Router) {
		r.Get("/{id}", galleriesC.Show)
		r.Get("/{id}/images/{filename}", galleriesC.Image)
		r.Group(func(r chi.Router) {
			r.Use(umw.RequireUser)
			r.Get("/", galleriesC.Index)
			r.Get("/new", galleriesC.New)
			r.Post("/", galleriesC.Create)
			r.Get("/{id}/edit", galleriesC.Edit)
			r.Post("/{id}", galleriesC.Update)
			r.Post("/{id}/delete", galleriesC.Delete)
			r.Post("/{id}/images", galleriesC.UploadImage)
			r.Post("/{id}/images/{filename}/delete", galleriesC.DeleteImage)
		})
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})
	//-----------------------------------------------------
	// Start the Server
	fmt.Printf("Starting server at port %s...\n", cfg.Server.Address)
	err = http.ListenAndServe(cfg.Server.Address, r)
	if err != nil {
		panic(err)
	}
	//once starting the router, all these methods are registered on the router
	//one receiving any request these methods are matched and the methods in the arguments gets executed for the response
}
