package main

import (
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"text/template"
)

// The type represent only one template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// method ServeHTTP handle HTTP requests
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.templ.Execute(w, r)
}

func main() {
	var addr = flag.String("addr", ":8080", "The addr of the application.")
	flag.Parse() // parse the flags

	r := newRoom()
	// main path
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)
	// run the room going
	go r.run()
	// start the web server
	log.Println("Starting web server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
