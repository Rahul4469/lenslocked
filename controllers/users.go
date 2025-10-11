package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Rahul4469/lenslocked/context"
	"github.com/Rahul4469/lenslocked/errors"
	"github.com/Rahul4469/lenslocked/models"
)

type Users struct {
	//Interface injection to implement Execute method
	Templates struct {
		New            Template
		SignIn         Template
		ForgotPassword Template
		CheckYourEmail Template
		ResetPassword  Template
	}
	UserService          *models.UserService
	SessionService       *models.SessionService
	PasswordResetService *models.PasswordResetService
	EmailService         *models.EmailService
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
	var data struct { // to pass as arg in New.Execute, previously there was no use
		Email    string
		Password string
	}
	data.Email = r.FormValue("email")
	data.Password = r.FormValue("password")
	user, err := u.UserService.Create(data.Email, data.Password)
	if err != nil {
		if errors.Is(err, models.ErrEmailTaken) {
			err = errors.Public(err, `That Email address is already associated 
			with an account.`)
		}
		u.Templates.New.Execute(w, r, data, err)
		return
	}
	//User created with email & password coming from the "Form"
	//and User details stored in user variable

	//Now exactly at the moment of user creation in db generate
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
	ctx := r.Context()
	user := context.User(ctx)
	// if user == nil {
	// 	http.Redirect(w, r, "/signin", http.StatusFound)
	// 	return
	// }
	//We dont need to check if user is nill, because at this point we
	//already would have checked the RequireUser ctx and have our user
	fmt.Fprintf(w, "Current user: %s\n", user.Email)

	// tokenCookie, err := r.Cookie("session")
	// if err != nil {
	// 	fmt.Println(err)
	// 	http.Redirect(w, r, "/signin", http.StatusFound)
	// 	return
	// }
	// //User method defined in Session model
	// //hashes token "value" from request -> checks and compares userId based on that from session table
	// // if the token hash matches -> returns user from that userId
	// user, err = u.SessionService.User(tokenCookie.Value)
	// if err != nil {
	// 	fmt.Fprint(w, "The Email cookie could not be read")
	// 	http.Redirect(w, r, "/signin", http.StatusFound)
	// 	return
	// }
	// fmt.Fprintf(w, "Current user: %s\n", user.Email)
}

// Delete Session token from DB & Set cookie to "" value
func (u Users) ProcessSignOut(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	err = u.SessionService.Delete(tokenCookie.Value)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	// Delete the user's cookie
	deleteCookie(w, tokenCookie.Value)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (u Users) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	u.Templates.ForgotPassword.Execute(w, r, data)
}

// on Button click(reset password)
// -> Create reset-token -> save it in pw-rest db -> send email with token -> render the check-your-email html page
func (u Users) ProcessForgotPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Email string
	}
	data.Email = r.FormValue("email")
	pwReset, err := u.PasswordResetService.Create(data.Email)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something WEnt wrong.", http.StatusNotFound)
		return
	}
	vals := url.Values{
		"token": {pwReset.Token},
	}
	resetURL := "https://www.lenslocked.com/reset-pw?" + vals.Encode()
	err = u.EmailService.ForgotPassword(data.Email, resetURL)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something WEnt wrong.", http.StatusNotFound)
		return
	}
	//Dont render the reset token here, we need the user to confirm they have
	//access to the email account to verify their identity
	u.Templates.CheckYourEmail.Execute(w, r, data)
}

func (u Users) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token string
	}
	data.Token = r.FormValue("token")
	u.Templates.ResetPassword.Execute(w, r, data)
}

func (u Users) ProcessResetPassword(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Token    string
		Password string
	}
	data.Token = r.FormValue("token")
	data.Password = r.FormValue("password")

	user, err := u.PasswordResetService.Consume(data.Token)
	if err != nil {
		fmt.Println(err)
		///TODO: distinguish bw types of errors.
		http.Error(w, "Something went wrong.", http.StatusNotFound)
		return
	}

	//Update the user's password.
	err = u.UserService.UpdatePassword(user.ID, data.Password)
	if err != nil {
		fmt.Println(err)
		///TODO: distinguish bw types of errors.
		http.Error(w, "Something went wrong.", http.StatusNotFound)
		return
	}

	//Sign the user in now that the password has been reset.
	//Any errors from this point onwards should redirect the user to the signin page.
	session, err := u.SessionService.Create(user.ID)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusFound)
		return
	}
	setCookie(w, CookieSession, session.Token)
	http.Redirect(w, r, "/users/me", http.StatusFound)
}

// ---------------------------------------------------------------------------
// Uses Session
type UserMiddleware struct {
	SessionService *models.SessionService
}

// - look up session data(name, value)
// - fetch User using the session
// - set base context to "request"
// - save User data to that context
// - pass the context to the request data,
// next implies that its gonna take the next http.Handler
// to call when we are done with the middleware
func (umw UserMiddleware) SetUser(next http.Handler) http.Handler {
	//HandlerFunc implements the http.Handler interface
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, err := r.Cookie("session")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		user, err := umw.SessionService.User(tokenCookie.Value)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (umw UserMiddleware) RequireUser(next http.Handler) http.Handler {
	//HandlerFunc implements the http.Handler interface
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/signin", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)

	})
}
