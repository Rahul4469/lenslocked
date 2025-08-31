package controllers

import (
	"net/http"

	"github.com/Rahul4469/lenslocked/views"
)

func StaticHandler(tpl views.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, nil)
	}
}

func FAQ(tpl views.Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   string
	}{
		{
			Question: "Is there a free version",
			Answer:   "Yes we offer a free trial for 30 days on any paid plans",
		},
		{
			Question: "Is there a free version",
			Answer:   "Yes we offer a free trial for 30 days on any paid plans",
		},
		{
			Question: "Is there a free version",
			Answer:   "Yes we offer a free trial for 30 days on any paid plans",
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, questions)
	}
}
