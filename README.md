# wsframe

**wsframe** is a Go-based websocket framework designed for building real-time web applications backed by a PostgreSQL database. It integrates HTTP routing, secure user authentication, JSON data storage, and dynamic templating with websocket communication.

---

## Features

- Websocket server handling real-time requests via [wsrooms](https://github.com/joncody/wsrooms)
- User authentication with salted password hashing and secure cookies
- Role-based route handling (`admin`, `authorized`, `public`)
- Dynamic route matching with regex support and JSON data retrieval from PostgreSQL
- Flexible templating system using Go's `html/template` with helper functions
- REST endpoints for login, registration, and logout
- Automatic table creation and JSON storage in PostgreSQL
- Static file serving and customizable routes

---

## Installation

Make sure you have Go installed (version 1.18+ recommended) and PostgreSQL running.

```bash
go get github.com/gorilla/mux
go get github.com/gorilla/securecookie
go get github.com/joncody/wsrooms
go get github.com/lib/pq
````

Clone this repo and build your application:

```bash
git clone https://github.com/joncody/wsframe.git
cd wsframe
go build -o wsframe
```

---

## Configuration

The app expects a JSON config file specifying database credentials, routes, templates, and security keys. Example config (`config.json`):

```json
{
  "name": "wsframe",
  "hashkey": "very-secret",
  "blockkey": "a-lotvery-secret",
  "port": "8080",
  "sslport": "0",
  "database": {
    "user": "dbuser",
    "password": "dbpass",
    "name": "dbname"
  },
  "routes": [
    {
      "route": "^/dashboard$",
      "admin": {
        "table": "admin_data",
        "key": "dashboard",
        "template": "admin_dashboard.tmpl",
        "controllers": "adminCtrl"
      },
      "authorized": {
        "privilege": "user,editor",
        "table": "user_data",
        "key": "dashboard",
        "template": "user_dashboard.tmpl",
        "controllers": "userCtrl"
      },
      "table": "public_data",
      "key": "dashboard",
      "template": "public_dashboard.tmpl",
      "controllers": "publicCtrl"
    }
  ]
}
```

---

## Usage

Initialize and start your app:

```go
package main

import (
  "log"
  "wsframe"
)

func main() {
  app := wsframe.NewApp("config.json")
  app.Start()
}
```

* The server listens on the configured port (default 8080).
* HTTP routes `/login`, `/register`, `/logout` handle user authentication.
* Websocket endpoint `/ws` manages real-time client-server communication.
* Static files are served from `/static/`.
* Dynamic routes are configured in the JSON file with route regex, templates, and data sources.

---

## Authentication

* Users register by sending a POST to `/register` with `alias` and `passhash` (SHA-1 hashed password).
* Passwords are salted and hashed with SHA-256 before storage.
* Login is done via POST `/login` with the same fields.
* Secure cookies store session info with `alias` and `privilege`.
* Privileges control route access (`admin`, `user`, etc.).

---

## Database

* Uses PostgreSQL with JSON storage.
* Automatically creates required tables (`auth` plus any specified in routes).
* Data is stored as JSON in a `value` column keyed by a unique `key`.
* Supports retrieving single rows or all rows from tables.

---

## Templates and Controllers

* Uses Go's `html/template` with helper functions for string and numeric operations.
* Templates are stored in `./static/views/`.
* Controllers are JavaScript modules loaded client-side to enhance views.
* The server sends rendered templates with the list of controllers over websocket.

---

## Extending the App

* Add new routes dynamically via `AddRoute` with custom handlers.
* Extend user privileges and authentication logic as needed.
* Customize templates and controllers to fit your frontend needs.
* Use websocket message handlers to build real-time interactive apps.

---

## License

See the [LICENSE](./LICENSE) file for details.
