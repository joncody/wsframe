package wsframe

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

type Auth struct {
	Alias     string `json:"alias,omitempty"`
	Passhash  string `json:"passhash"`
	Salt      string `json:"salt"`
	Hash      string `json:"hash"`
	Privilege string `json:"privilege"`
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
		Domain:   "localhost",
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

func (wfa *App) register(w http.ResponseWriter, r *http.Request) {
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

func (wfa *App) login(w http.ResponseWriter, r *http.Request) {
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

func (wfa *App) logout(w http.ResponseWriter, r *http.Request) {
	wfa.SetCookie(w, r, nil, true)
	w.WriteHeader(http.StatusOK)
}
