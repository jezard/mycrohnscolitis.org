//handle the main functions
package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"

	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/jezard/mycrohnscolitis.org/conf"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
	"github.com/markbates/goth/providers/twitter"
)

var config = conf.Configuration()

//Compile templates on start (http://sanatgersappa.blogspot.co.uk/2013/11/creating-master-page-for-your-go-web-app.html)
var templates = template.Must(template.ParseFiles(config.Tpath+"reused/header.html", config.Tpath+"reused/footer.html", config.Tpath+"home.html", config.Tpath+"about.html", config.Tpath+"login.html", config.Tpath+"user.html"))

//Page - content to be passed to page
type Page struct {
	Title        string
	Providers    []string
	ProvidersMap map[string]string
	User         goth.User
}

func init() {
	gothic.Store = sessions.NewFilesystemStore(os.TempDir(), []byte("goth-example"))
}

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl, data)
}

// SEE: https://www.socketloop.com/tutorials/golang-gorilla-mux-routing-example
func main() {
	goth.UseProviders(
		twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://www.mycrohnscolitis.org:8080/auth/twitter/callback"),
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"), "http://www.mycrohnscolitis.org:8080/auth/gplus/callback"), //https://console.developers.google.com/apis/credentials/wizard?api=plus-json.googleapis.com&project=mycrohnscolitis&authuser=1
	)
	m := make(map[string]string)
	m["twitter"] = "Twitter"
	m["gplus"] = "Google Plus"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	p := pat.New()

	p.Get("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {

		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		//t, _ := template.New("foo").Parse(userTemplate)
		//t.Execute(w, user)
		fmt.Printf("User: %#v", user) //all the information is stored in $user
		display(w, "user", &Page{Title: "User Page", User: user})
	})

	p.Get("/auth/{provider}", gothic.BeginAuthHandler)
	p.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Login")
		display(w, "login", &Page{Title: "Login Page", Providers: keys, ProvidersMap: m})
	})

	fmt.Println("works")
	p.HandleFunc("/", HomeHandler)
	p.HandleFunc("/about", AboutHandler)
	// r.HandleFunc("/login", LoginHandler)
	// r.HandleFunc("/view-user", ViewUserHandler) //test page
	// r.HandleFunc("/diary/new/", DiaryNewHandler)
	// r.HandleFunc("/diary/view/", DiaryViewHandler)
	// r.HandleFunc("/diary/edit/", DiaryEditHandler)
	http.ListenAndServe(":8080", p)
}

//HomeHandler - do homepage stuff
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Home Page")
	display(w, "home", &Page{Title: "Home Page"})
}

//AboutHandler - do about page stuff
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("About Page")
	display(w, "about", &Page{Title: "About Page"})
}
