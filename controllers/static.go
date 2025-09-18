package controllers

import (
	"html/template"
	"net/http"
)

// Created to separate parse and Execute steps to router
func StaticHandler(tpl Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, nil)
	}
}

func FAQ(tpl Template) http.HandlerFunc {
	questions := []struct {
		Question string
		Answer   template.HTML
	}{
		{
			Question: "Is there a free version?",
			Answer:   "Yes we offer a free trial for 30 days on any paid plans",
		},
		{
			Question: "What are your support hourse?",
			Answer:   "We have a supoort team available 24/7 over email, response may take longer on weekends",
		},
		{
			Question: "How do I contact support?",
			Answer:   `Email us at <a href = "mailtosupport@gmail.com">support@app.com</a>`,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tpl.Execute(w, r, questions)
	}
}
