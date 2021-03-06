//handle the main functions
package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sort"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/pat"
	"github.com/gorilla/sessions"
	"github.com/jezard/mycrohnscolitis.org/conf"
	"github.com/jezard/mycrohnscolitis.org/diary"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/gplus"
	"github.com/markbates/goth/providers/twitter"
	//"github.com/markbates/goth/providers/twitter"
)

var config = conf.Configuration()

//Compile templates on start (http://sanatgersappa.blogspot.co.uk/2013/11/creating-master-page-for-your-go-web-app.html)
var templates = template.Must(template.ParseFiles(config.Tpath+"reused/header.html", config.Tpath+"reused/footer.html", config.Tpath+"home.html", config.Tpath+"about.html", config.Tpath+"login.html", config.Tpath+"user.html", config.Tpath+"diary-overview.html"))

//Page - content to be passed to page
type Page struct {
	Title        string
	Providers    []string
	ProvidersMap map[string]string
	User         goth.User
	ValidUser    bool
	Overview     diary.Overview
}

func init() {
	gothic.Store = sessions.NewFilesystemStore(os.TempDir(), []byte("authuser"))
}

var store = sessions.NewFilesystemStore(os.TempDir(), []byte("userIdentity"))

//Display the named template
func display(w http.ResponseWriter, tmpl string, data interface{}) {
	templates.ExecuteTemplate(w, tmpl, data)
}

// SEE: https://www.socketloop.com/tutorials/golang-gorilla-mux-routing-example
func main() {

	db, err := sql.Open("mysql", config.MySQLUser+":"+config.MySQLPass+"@tcp("+config.MySQLHost+":3306)/"+config.MySQLDB)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	goth.UseProviders(
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), "http://www.mycrohnscolitis.org:8080/auth/facebook/callback"), //https://developers.facebook.com/apps/1136643983096464/dashboard/
		twitter.New(os.Getenv("TWITTER_KEY"), os.Getenv("TWITTER_SECRET"), "http://www.mycrohnscolitis.org:8080/auth/twitter/callback"),
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"), "http://www.mycrohnscolitis.org:8080/auth/gplus/callback"), //https://console.developers.google.com/apis/credentials/wizard?api=plus-json.googleapis.com&project=mycrohnscolitis&authuser=1
	)
	m := make(map[string]string)
	m["facebook"] = "Facebook"
	m["gplus"] = "Google Plus"
	m["twitter"] = "Twitter"

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//handle the various page requests
	p := pat.New()
	p.Get("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}

		//save / update user to MySQL database
		err = login(db, user) //retrieve the local user id that matches our twitter or google ID
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
		}

		mysession, _ := store.Get(r, "userIdentity")
		mysession.Values["accessToken"] = user.AccessToken
		mysession.Save(r, w)

		display(w, "user", &Page{Title: "User Page", User: user, ValidUser: true})
	})
	p.Get("/auth/{provider}", gothic.BeginAuthHandler)
	p.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		display(w, "login", &Page{Title: "Login Page", Providers: keys, ProvidersMap: m})
	})
	p.HandleFunc("/", HomeHandler)
	p.HandleFunc("/about", AboutHandler)
	p.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		//delete our session
		session, _ := store.Get(r, "userIdentity")
		session.Options.MaxAge = -1
		session.Save(r, w)
		display(w, "home", &Page{Title: "Home Page", ValidUser: false})
	})
	p.Get("/diary/overview", diaryOverviewHandler)

	//serve the static resource files from /
	p.PathPrefix("/").Handler(http.FileServer(http.Dir(config.Rpath)))
	http.Handle("/", p)

	http.ListenAndServe(":8080", p)
}

//HomeHandler - do homepage stuff
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	_, isValid := ValidateUser(w, r)

	//snippet showing how we can get values directly from the session if required
	// session, _ := gothic.Store.Get(r, "authuser")
	// s := session.Values["user"]
	// user, _ := s.(goth.User)
	// fmt.Printf("session %s\n", user.AccessToken)

	display(w, "home", &Page{Title: "Home Page", ValidUser: isValid})
}

func diaryOverviewHandler(w http.ResponseWriter, r *http.Request) {
	_, isValid := ValidateUser(w, r)
	display(w, "diary-overview", &Page{Title: "Diary Overview", ValidUser: isValid, Overview: diary.GetOverview()})
}

//AboutHandler - do about page stuff
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	_, isValid := ValidateUser(w, r)
	display(w, "about", &Page{Title: "About Page", ValidUser: isValid})
}

func login(db *sql.DB, user goth.User) (err error) {
	_, err = db.Query("INSERT INTO user (auth_userid, auth_provider, access_token, name, nickname, avatar_url, last_login) VALUES (?,?,?,?,?,?, NOW()) ON DUPLICATE KEY UPDATE auth_userid=?, auth_provider=?, access_token=?, name=?, nickname=?, avatar_url=?, last_login=NOW()", user.UserID, user.Provider, user.AccessToken, user.Name, user.NickName, user.AvatarURL, user.UserID, user.Provider, user.AccessToken, user.Name, user.NickName, user.AvatarURL) //these inputs repeat once to match

	return
}

//ValidateUser return the user_id for use in queries and bool for hiding / showing in templates
func ValidateUser(w http.ResponseWriter, r *http.Request) (id int, validUser bool) {

	validUser = false

	db, err := sql.Open("mysql", config.MySQLUser+":"+config.MySQLPass+"@tcp("+config.MySQLHost+":3306)/"+config.MySQLDB)
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	mysession, _ := store.Get(r, "userIdentity")
	a := mysession.Values["accessToken"]
	accessToken, _ := a.(string)

	id = 0

	_ = db.QueryRow("SELECT id FROM user WHERE access_token = ? LIMIT 1", accessToken).Scan(&id)

	if id != 0 {
		validUser = true
	} else {
		//delete our session
		session, _ := store.Get(r, "userIdentity")
		session.Options.MaxAge = -1
		session.Save(r, w)
	}
	return
}
