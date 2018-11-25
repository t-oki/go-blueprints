package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"github.com/t-oki/go-blueprints/chapter2/trace"
)

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ =
			template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}
	if err := t.templ.Execute(w, data); err != nil {
		log.Fatal("template execute:", err)
	}
}

func main() {
	var addr = flag.String("addr", ":8080", "port to listen")
	flag.Parse()

	gomniauth.SetSecurityKey("security_key")
	gomniauth.WithProviders(
		// facebook.New("302045812422-09cv8f673asng4tst0k3u53a3ppu4h08.apps.googleusercontent.com","uPXmfGhxmCgGXN-OKrFwFK1I", "http://localhost:8080/auth/callback/facebook"),
		// github.New("302045812422-09cv8f673asng4tst0k3u53a3ppu4h08.apps.googleusercontent.com", "uPXmfGhxmCgGXN-OKrFwFK1I", "http://localhost:8080/auth/callback/github"),
		google.New("302045812422-09cv8f673asng4tst0k3u53a3ppu4h08.apps.googleusercontent.com", "uPXmfGhxmCgGXN-OKrFwFK1I", "http://localhost:8080/auth/callback/google"),
	)

	r := newRoom()
	r.tracer = trace.New(os.Stdout)
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))
	http.Handle("/login", &templateHandler{filename: "login.html"})
	http.HandleFunc("/auth/", loginHandler)
	http.Handle("/room", r)

	go r.run()
	log.Printf("Start Web Server on Port %s", *addr)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
