package wsframe

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	_ "github.com/lib/pq"
	"github.com/joncody/wsrooms"
)

type App struct {
	Name         string `json:"name"`
	HashKey      string `json:"hashkey"`
	BlockKey     string `json:"blockkey"`
	SecureCookie *securecookie.SecureCookie
	Templates    *template.Template
	Port         string `json:"port"`
	SSLPort      string `json:"sslport"`
	Database     DBConfig `json:"database"`
	Driver *sql.DB
	Routes []Route `json:"routes"`
	Added []AddedRoute
	Router *mux.Router
}

func (wfa *App) Start() {
	var err error

	wfa.SecureCookie = securecookie.New([]byte(wfa.HashKey), []byte(wfa.BlockKey))
	wfa.Driver, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", wfa.Database.User, wfa.Database.Password, wfa.Database.Name))
	if err == nil && wfa.Driver != nil {
		defer wfa.Driver.Close()
		wfa.prepareTables()
	}
	if wfa.SSLPort != "0" {
		go func() {
			log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%s", wfa.SSLPort), "server.crt", "server.key", wfa.Router))
		}()
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", wfa.Port), wfa.Router))
}

func NewApp(cj string) *App {
	app := &App{
		Name:         "wsframe",
		SecureCookie: nil,
		Templates:    template.Must(template.New("").Funcs(TemplateFuncs).ParseGlob("./static/views/*")),
		HashKey:      "very-secret",
		BlockKey:     "a-lotvery-secret",
		Port:         "8080",
		SSLPort:      "0",
		Database: DBConfig{
			User:     "dbuser",
			Password: "dbpass",
			Name:     "dbname",
		},
		Driver: nil,
		Routes: []Route{},
		Added: []AddedRoute{},
		Router: mux.NewRouter().StrictSlash(false),
	}
	cjb, err := ioutil.ReadFile(cj)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(cjb, &app); err != nil {
		log.Fatal(err)
	}
	app.Router.HandleFunc("/login", app.login).Methods("POST")
	app.Router.HandleFunc("/register", app.register).Methods("POST")
	app.Router.HandleFunc("/logout", app.logout).Methods("POST")
	app.Router.HandleFunc("/ws", wsrooms.SocketHandler(app.ReadCookie)).Methods("GET")
	app.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	app.Router.PathPrefix("/").Handler(http.HandlerFunc(app.baseHandler)).Methods("GET")
	wsrooms.Emitter.On("request", app.processRequest)
	return app
}
