package wsframe

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/joncody/wsrooms"
)

type RouteConfig struct {
	Table       string `json:"table"`
	Key         string `json:"key"`
	Template    string `json:"template"`
	Controllers string `json:"controllers"`
	Privilege   string `json:"privilege,omitempty"` // Only used for Authorized routes
}

type Route struct {
	Route      string      `json:"route"`
	Admin      RouteConfig `json:"admin"`
	Authorized RouteConfig `json:"authorized"`
	General    RouteConfig `json:"general"`
}

type AddedRoute struct {
	Route   string
	Handler func(c *wsrooms.Conn, msg *wsrooms.Message, matches []string)
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

func (wfa *App) AddRoute(path string, handler func(c *wsrooms.Conn, msg *wsrooms.Message, matches []string)) {
	route := &AddedRoute{
		Route:   path,
		Handler: handler,
	}
	wfa.Added = append(wfa.Added, route)
}

func (wfa *App) baseHandler(w http.ResponseWriter, r *http.Request) {
	cook := wfa.ReadCookie(r)
	wfa.Templates.ExecuteTemplate(w, "base", cook)
}

func (wfa *App) Render(c *wsrooms.Conn, msg *wsrooms.Message, template string, controllers []string, data interface{}) {
	var tpl bytes.Buffer

	if err := wfa.Templates.ExecuteTemplate(&tpl, template, data); err != nil {
		log.Println(err)
	}
	resp := struct {
		Template    string   `json:"template"`
		Controllers []string `json:"controllers"`
	}{
		Template:    tpl.String(),
		Controllers: controllers,
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
}

func (wfa *App) processRequest(c *wsrooms.Conn, msg *wsrooms.Message) {
	var data interface{}
	var controllers []string
	var table, key, template, ctrls string

	path := string(msg.Payload)
	for _, added := range wfa.Added {
		pattern := regexp.MustCompile(added.Route)
		if pattern.MatchString(path) == true {
			subs := pattern.FindStringSubmatch(path)
			added.Handler(c, msg, subs)
			return
		}
	}
	for _, details := range wfa.Routes {
		pattern := regexp.MustCompile(details.Route)
		if pattern.MatchString(path) == false {
			continue
		}
		if c.Cookie["privilege"] == "admin" && (details.Admin.Template != "" || details.Admin.Controllers != "") {
			table = details.Admin.Table
			key = details.Admin.Key
			template = details.Admin.Template
			ctrls = details.Admin.Controllers
		} else if c.Cookie["privilege"] != "" && details.Authorized.Privilege != "" && strings.Contains(details.Authorized.Privilege, c.Cookie["privilege"]) {
			table = details.Authorized.Table
			key = details.Authorized.Key
			template = details.Authorized.Template
			ctrls = details.Authorized.Controllers
		} else {
			table = details.Table
			key = details.Key
			template = details.Template
			ctrls = details.Controllers
		}
		if table != "" {
			if strings.HasPrefix(table, "$") {
				subs := pattern.FindStringSubmatch(path)
				tablenum, err := strconv.Atoi(string(table[1]))
				if err != nil {
					log.Println(err)
				} else if len(subs) >= tablenum {
					table = subs[tablenum]
				}
			}
			if strings.HasPrefix(key, "$") {
				subs := pattern.FindStringSubmatch(path)
				keynum, err := strconv.Atoi(string(key[1]))
				if err != nil {
					log.Println(err)
				} else if len(subs) >= keynum {
					key = subs[keynum]
				}
			}
			if key != "" {
				data = wfa.GetRow(table, key)
			} else {
				data = wfa.GetRows(table)
			}
		}
		controllers = strings.Split(ctrls, ",")
		wfa.Render(c, msg, template, controllers, data)
		break
	}
}