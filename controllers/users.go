package controllers

import (
	"fmt"
	"net/http"

	"github.com/Rahul4469/lenslocked/models"
)

type Users struct {
	//Interface injection to implement Execute method
	Templates struct {
		New    Template
		SignIn Template
	}
	UserService *models.UserService
}

// to dispaly signup form
func (u Users) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
		// CSRFField template.HTML ----added csrf managing logic to views.ParseFS
	}

	data.Email = r.FormValue("email")
	// data.CSRFField = csrf.TemplateField(r)
	u.Templates.New.Execute(w, r, data) //Note: any data/field/variable thats going through Execute method can be rendered on the html page with {{.__}}
}

func (u Users) Create(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := u.UserService.Create(email, password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User Created: %+v", user)
}

func (u Users) SignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	//Using email to input pre filled Email field
	//on the signup form
	data.Email = r.FormValue("email")
	u.Templates.SignIn.Execute(w, r, data)
}

func (u Users) ProcessSignIn(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Authenticate(data.Email, data.Password)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "email",
		Value:    user.Email,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)
	fmt.Fprintf(w, "User authenticated: %+v", user)

}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	email, err := r.Cookie("email")
	if err != nil {
		fmt.Fprint(w, "The Email cookie could not be read")
		return
	}
	fmt.Fprintf(w, "Email cookie: %s", email.Value)
}
