package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jezard/mycrohnscolitis.org/conf"
)

var config = conf.Configuration()

//Compile templates on start (http://sanatgersappa.blogspot.co.uk/2013/11/creating-master-page-for-your-go-web-app.html)
var templates = template.Must(template.ParseFiles(config.Tpath+"reused/header.html", config.Tpath+"reused/footer.html", config.Tpath+"home.html", config.Tpath+"about.html"))

type Page struct {
	Title string
}

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl, data)
}

// SEE: https://www.socketloop.com/tutorials/golang-gorilla-mux-routing-example
func main() {
	r := mux.NewRouter()
	fmt.Println("works")
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/about", AboutHandler)
	// r.HandleFunc("/diary/new/", DiaryNewHandler)
	// r.HandleFunc("/diary/view/", DiaryViewHandler)
	// r.HandleFunc("/diary/edit/", DiaryEditHandler)
	http.ListenAndServe(":8080", r)
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//comment
	fmt.Println("Home Page")
	display(w, "home", &Page{Title: "Home Page"})
}
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	//comment
	fmt.Println("About Page")
	display(w, "about", &Page{Title: "About Page"})
}
