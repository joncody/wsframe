package wsframe

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/joncody/wsrooms"
	_ "github.com/lib/pq"
)

type App struct {
	Name         string `json:"name"`
	HashKey      string `json:"hashkey"`
	BlockKey     string `json:"blockkey"`
	SecureCookie *securecookie.SecureCookie
	Templates    *template.Template
	Port         string `json:"port"`
	SSLPort      string `json:"sslport"`
	Database     struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	} `json:"database"`
	Driver *sql.DB
	Routes map[string]struct {
		Table       string `json:"table"`
		Key         string `json:"key"`
		Template    string `json:"template"`
		Controllers string `json:"controllers"`
	} `json:"routes"`
	Router *mux.Router
}

type Auth struct {
	Alias     string `json:"alias,omitempty"`
	Passhash  string `json:"passhash"`
	Salt      string `json:"salt"`
	Hash      string `json:"hash"`
	Privilege string `json:"privilege"`
}

func ToKey(s string) string {
	respecial := regexp.MustCompile(`([^a-z0-9_\-\s]+)`)
	s = strings.ToLower(s)
	s = respecial.ReplaceAllString(s, "")
	s = strings.Replace(s, " - ", "_-_", -1)
	s = strings.Replace(s, " ", "-", -1)
	return strings.Trim(s, "-")
}

func FromKey(s string) string {
	s = strings.Replace(s, "-", " ", -1)
	s = strings.Replace(s, "_ _", " - ", -1)
	s = strings.Title(s)
	return strings.Trim(s, " ")
}

var TemplateFuncs = template.FuncMap{
	"unescaped": func(x string) interface{} {
		return template.HTML(x)
	},
	"sha1sum": func(x string) string {
		return fmt.Sprintf("%x", sha1.Sum([]byte(x)))
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"add": func(a, b int) int {
		return a + b
	},
	"multiply": func(a, b int) int {
		return a * b
	},
	"divide": func(a, b int) int {
		return a / b
	},
	"usd": func(x int) string {
		return fmt.Sprintf("$%.2f", float64(x)/float64(100))
	},
	"css": func(s string) template.CSS {
		return template.CSS(s)
	},
	"tokey":   ToKey,
	"fromkey": FromKey,
}

func (wfa *App) ReadCookie(r *http.Request) map[string]string {
	value := make(map[string]string)
	cookie, err := r.Cookie(wfa.Name)
	if err != nil {
		return value
	}
	if err := wfa.SecureCookie.Decode(wfa.Name, cookie.Value, &value); err != nil {
		return value
	}
	return value
}

func (wfa *App) SetCookie(w http.ResponseWriter, r *http.Request, value map[string]string, logout bool) {
	encoded, err := wfa.SecureCookie.Encode(wfa.Name, value)
	if err != nil {
		return
	}
	cookie := &http.Cookie{
		Name:     wfa.Name,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
	}
	if logout == true {
		cookie.Expires = time.Now().Add(-24 * time.Hour)
		cookie.MaxAge = -1
	} else {
		cookie.Expires = time.Now().Add(24 * time.Hour)
		cookie.MaxAge = 60 * 60 * 24
	}
	http.SetCookie(w, cookie)
	return
}

func (wfa *App) Register(w http.ResponseWriter, r *http.Request) {
	var err error

	alias := r.FormValue("alias")
	passhash := r.FormValue("passhash")
	value := make([]byte, 0)
	row := wfa.Driver.QueryRow(`SELECT value FROM auth WHERE key = $1`, alias)
	if err = row.Scan(&value); err != sql.ErrNoRows {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	randombytes := make([]byte, 16)
	if _, err = rand.Read(randombytes); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	salt := fmt.Sprintf("%x", sha1.Sum(randombytes))
	auth := Auth{
		Passhash:  passhash,
		Salt:      salt,
		Privilege: "user",
		Hash:      fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s%s%s", alias, passhash, salt)))),
	}
	if value, err = json.Marshal(&auth); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if _, err = wfa.Driver.Exec(fmt.Sprintf(`INSERT INTO auth (key, value) VALUES ($1, $2)`), alias, value); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		wfa.SetCookie(w, r, map[string]string{
			"alias":     alias,
			"privilege": auth.Privilege,
		}, false)
		w.WriteHeader(http.StatusOK)
	}
}

func (wfa *App) Login(w http.ResponseWriter, r *http.Request) {
	alias := r.FormValue("alias")
	passhash := r.FormValue("passhash")
	value := make([]byte, 0)
	row := wfa.Driver.QueryRow(`SELECT value FROM auth WHERE key = $1`, alias)
	if err := row.Scan(&value); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	auth := Auth{}
	if err := json.Unmarshal(value, &auth); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s%s%s", alias, passhash, auth.Salt))))
	if hash == auth.Hash {
		wfa.SetCookie(w, r, map[string]string{
			"alias":     alias,
			"privilege": auth.Privilege,
		}, false)
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (wfa *App) Logout(w http.ResponseWriter, r *http.Request) {
	wfa.SetCookie(w, r, nil, true)
	w.WriteHeader(http.StatusOK)
}

func (wfa *App) GetRow(table, key string) string {
	value := make([]byte, 0)
	row := wfa.Driver.QueryRow(fmt.Sprintf(`SELECT value FROM %s WHERE key = $1`, table), key)
	if err := row.Scan(&value); err != nil {
		log.Println(err)
		return ""
	}
	return string(value)
}

func (wfa *App) GetRows(table string) []string {
	values := make([]string, 0)
	rows, err := wfa.Driver.Query(fmt.Sprintf(`SELECT value FROM %s`, table))
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		value := make([]byte, 0)
		if err := rows.Scan(&value); err != nil {
			log.Println(err)
			return values
		}
		values = append(values, string(value))
	}
	return values
}

func (wfa *App) InsertRow(table, key, value string) error {
	if _, err := wfa.Driver.Exec(fmt.Sprintf(`INSERT INTO %s (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, table), key, value); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (wfa *App) ProcessRequest(c *wsrooms.Conn, msg *wsrooms.Message) {
	var tpl bytes.Buffer
	var data interface{}

	path := string(msg.Payload)
	for route, details := range wfa.Routes {
		pattern := regexp.MustCompile(route)
		if pattern.MatchString(path) == false {
			continue
		}
		if details.Key != "" && details.Table != "" {
			data = wfa.GetRow(details.Table, details.Key)
		} else if details.Table != "" {
			data = wfa.GetRows(details.Table)
		}
		if err := wfa.Templates.ExecuteTemplate(&tpl, details.Template, data); err != nil {
			log.Println(err)
		}
		resp := struct {
			Template    string   `json:"template"`
			Controllers []string `json:"controllers"`
		}{
			Template:    tpl.String(),
			Controllers: strings.Split(details.Controllers, ","),
		}
		payload, err := json.Marshal(&resp)
		if err != nil {
			log.Println(err)
			return
		}
		msg.EventLength = len("response")
		msg.Event = "response"
		msg.PayloadLength = len(payload)
		msg.Payload = payload
		c.Send <- msg.Bytes()
		break
	}
}

func (wfa *App) BaseHandler(w http.ResponseWriter, r *http.Request) {
	cook := wfa.ReadCookie(r)
	wfa.Templates.ExecuteTemplate(w, "base", cook)
}

func (wfa *App) PrepareTables() {
	const query = `CREATE TABLE IF NOT EXISTS %s (
				       id bigserial PRIMARY KEY,
					   key text UNIQUE NOT NULL,
					   value json
                  )`

	tables := []string{"auth"}
	for _, details := range wfa.Routes {
		if details.Table == "" {
			continue
		}
		tables = append(tables, details.Table)
	}
	for _, table := range tables {
		if _, err := wfa.Driver.Exec(fmt.Sprintf(query, table)); err != nil {
			log.Fatal(err)
		}
	}
}

func (wfa *App) Start() {
	var err error

	wfa.SecureCookie = securecookie.New([]byte(wfa.HashKey), []byte(wfa.BlockKey))
	wfa.Driver, err = sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", wfa.Database.User, wfa.Database.Password, wfa.Database.Name))
	if err == nil && wfa.Driver != nil {
		defer wfa.Driver.Close()
		wfa.PrepareTables()
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
		BlockKey:     "a-lot-secret",
		Port:         "8080",
		SSLPort:      "0",
		Database: struct {
			User     string `json:"user"`
			Password string `json:"password"`
			Name     string `json:"name"`
		}{
			User:     "dbuser",
			Password: "dbpass",
			Name:     "dbname",
		},
		Driver: nil,
		Routes: map[string]struct {
			Table       string `json:"table"`
			Key         string `json:"key"`
			Template    string `json:"template"`
			Controllers string `json:"controllers"`
		}{},
		Router: mux.NewRouter().StrictSlash(false),
	}
	cjb, err := ioutil.ReadFile(cj)
	if err != nil {
		log.Fatal(err)
	}
	if err = json.Unmarshal(cjb, &app); err != nil {
		log.Fatal(err)
	}
	app.Router.HandleFunc("/login", app.Login).Methods("POST")
	app.Router.HandleFunc("/register", app.Register).Methods("POST")
	app.Router.HandleFunc("/logout", app.Logout).Methods("POST")
	app.Router.HandleFunc("/ws", wsrooms.SocketHandler(app.ReadCookie)).Methods("GET")
	app.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	app.Router.PathPrefix("/").Handler(http.HandlerFunc(app.BaseHandler)).Methods("GET")
	wsrooms.Emitter.On("request", app.ProcessRequest)
	return app
}
