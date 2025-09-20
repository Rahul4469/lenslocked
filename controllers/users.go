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
	UserService    *models.UserService
	SessionService *models.SessionService
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
	//User created with email & password coming from Form
	//and User details stored in user variable

	//Now exaclty at the moment of user creation generate
	//session token and save to sessions table
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		//long term we should show a warning about not being able to sign the user in
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	// cookie := http.Cookie{
	// 	Name:     "session",
	// 	Value:    session.Token,
	// 	Path:     "/",
	// 	HttpOnly: true,
	// }
	// http.SetCookie(w, &cookie)
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)

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
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "users/me", http.StatusFound)

}

func (u Users) CurrentUser(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("session")
	if err != nil {
		fmt.Fprint(w, "The Email cookie could not be read")
		http.Redirect(w, r, "signin", http.StatusFound)
		return
	}
	//User method defined in Session model
	//hashes token "value" from request -> checks and compares userId based on that from session table
	// if the token hash matches -> returns user from that userId
	user, err := u.SessionService.User(tokenCookie.Value)
	if err != nil {
		fmt.Fprint(w, "The Email cookie could not be read")
		http.Redirect(w, r, "signin", http.StatusFound)
		return
	}
	fmt.Fprintf(w, "Current user: %s\n", user.Email)
}
