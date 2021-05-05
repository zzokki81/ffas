package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-mail/mail"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Account struct {
	Username  string
	Password  string
	Host      string
	Port      int
	FromEmail string
	ToEmail   string
}
type Message struct {
	Name    string
	Email   string
	Subject string
	Content string
	Errors  map[string]string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	r.HandleFunc("/confirm", confirm).Methods("GET")
	r.HandleFunc("/", contact).Methods("GET")
	r.HandleFunc("/", send).Methods("POST")

	log.Println("Listening at port 8080....")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func contact(w http.ResponseWriter, r *http.Request) {
	render(w, "static/templates/index.html", nil)
}

func confirm(w http.ResponseWriter, r *http.Request) {
	render(w, "static/templates/confirmation.html", nil)
}

func send(w http.ResponseWriter, r *http.Request) {
	msg := &Message{
		Name:    r.FormValue("name"),
		Email:   r.FormValue("email"),
		Subject: r.FormValue("subject"),
		Content: r.FormValue("content"),
	}

	if !msg.Validate() {
		render(w, "static/templates/index.html", msg)
		return
	}
	if err := msg.Deliver(); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/confirm", http.StatusSeeOther)
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

	if strings.TrimSpace(msg.Name) == "" {
		msg.Errors["Name"] = "Please input a name"
	}

	if strings.TrimSpace(msg.Content) == "" {
		msg.Errors["Content"] = "Please input a message"
	}

	return len(msg.Errors) == 0
}

func (msg *Message) Deliver() error {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		panic(err)
	}
	config := Account{
		Username:  os.Getenv("ACCOUNT_USERNAME"),
		Password:  os.Getenv("ACCOUNT_PASSWORD"),
		Host:      os.Getenv("SMTP_HOST"),
		Port:      port,
		FromEmail: os.Getenv("FROM_EMAIL"),
		ToEmail:   os.Getenv("TO_EMAIL"),
	}
	messageBody := fmt.Sprintf("%v\n\nContact email: \n%v", msg.Content, msg.Email)

	email := mail.NewMessage()
	email.SetHeader("To", config.ToEmail)
	email.SetHeader("From", config.FromEmail)
	email.SetHeader("Reply-To", msg.Email)
	email.SetHeader("Subject", msg.Subject)
	email.SetBody("text/plain", messageBody)

	return mail.NewDialer(config.Host, config.Port, config.Username, config.Password).DialAndSend(email)
}
