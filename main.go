package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
)

type Message struct {
	Name    string
	Email   string
	Subject string
	Content string
	Errors  map[string]string
}

func main() {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/", index).Methods("GET")
	log.Println("Listening at port :8080....")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func index(w http.ResponseWriter, r *http.Request) {
	render(w, "static/templates/index.html", nil)
}

func render(w http.ResponseWriter, filename string, data interface{}) {
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (msg *Message) Validate() bool {
	msg.Errors = make(map[string]string)

	re := regexp.MustCompile(".+@.+\\..+")
	matched := re.Match([]byte(msg.Email))
	if !matched {
		msg.Errors["Email"] = "Invalid email"
	}

	if strings.TrimSpace(msg.Content) == "" {
		msg.Errors["Name"] = "Please input a name"
	}

	if strings.TrimSpace(msg.Content) == "" {
		msg.Errors["Content"] = "Please input a message"
	}

	return len(msg.Errors) == 0
}
