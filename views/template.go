package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/Rahul4469/cloud-memory/context"
	"github.com/Rahul4469/cloud-memory/models"
	"github.com/gorilla/csrf"
)

type public interface {
	Public() string
}

// Parse html template and save into tpl
func ParseFS(fs fs.FS, patterns ...string) (Template, error) {
	tpl := template.New(path.Base(patterns[0]))
	// Injecting custom functions into the html templates
	tpl.Funcs(
		template.FuncMap{
			//We Implement this Func down in execute becuse we needs request specific info
			//that could could be done in Execute where we passed r *request
			"csrfField": func() (template.HTML, error) {
				return "", fmt.Errorf("CSRFField not implemented")
			},
			"currentUser": func() (template.HTML, error) {
				return "", fmt.Errorf("current user not implemented")
			},
			"errors": func() []string {
				return nil
			},
		},
	)
	tpl, err := tpl.ParseFS(fs, patterns...)
	if err != nil {
		return Template{}, fmt.Errorf("parsing template: %w", err)
	}
	return Template{htmlTpl: tpl}, nil
}

// func Parse(filepath string) (Template, error) {
// 	tpl, err := template.ParseFiles(filepath)
// 	if err != nil {
// 		return Template{}, fmt.Errorf("parsing template: %w", err)
// 	}
// 	return Template{htmlTpl: tpl}, nil
// }

type Template struct {
	htmlTpl *template.Template
}

// helper func to reuse for templates
// Execute writes the tpl data as a response to the client
func (t Template) Execute(w http.ResponseWriter, r *http.Request, data interface{}, errs ...error) {
	tpl, err := t.htmlTpl.Clone() //Clone(): To allow multiple requests without overlapping data
	if err != nil {
		log.Printf("Cloning Template: %v", err)
		http.Error(w, "There was and error rendering the page.", http.
			StatusInternalServerError)
		return
	}

	//So this error implementation will only work if you wrap your errors
	//with Public() method wherever you are handling your errors
	errMsgs := errMessages(errs...)
	//Registers custom functions that can be called from within your HTML templates
	//Funcs must be called before ParseFS/Parse
	tpl.Funcs(
		template.FuncMap{
			"csrfField": func() template.HTML {
				return csrf.TemplateField(r)
			},
			"currentUser": func() *models.User {
				return context.User(r.Context())
			},
			"errors": func() []string {
				return errMsgs
				// var errMessages []string
				// for _, err := range errs {
				// 	var pubError public
				// 	if errors.As(err, &pubError) {
				// 		errMessages = append(errMessages, pubError.Public())
				// 	} else {
				// 		errMessages = append(errMessages, "Something went wrong")
				// 	}
				// }
				// return errMessages
			},
		},
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//err := t.htmlTpl.Execute(w, data) ---Before writing CSRF code
	var buf bytes.Buffer //
	err = tpl.Execute(&buf, data)
	if err != nil {
		log.Printf("parsing template: %v", err)
		http.Error(w, "There was an error Executing the template", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf) //
}

func errMessages(errs ...error) []string {
	var errMessages []string
	for _, err := range errs {
		var pubError public
		if errors.As(err, &pubError) {
			errMessages = append(errMessages, pubError.Public())
		} else {
			errMessages = append(errMessages, "Something went wrong")
		}
	}
	return errMessages
}
