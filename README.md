# 🧩 WSFrame

**WSFrame** is a lightweight Go framework for real-time, dynamic web applications using WebSockets, templates, and PostgreSQL. It is designed to work with [wsrooms](https://github.com/joncody/wsrooms) for bi-directional client-server communication.

---

## 🚀 Features

- 📡 WebSocket routing via regular expressions
- 🔄 Dynamic HTML rendering via Go templates
- 🔐 Built-in user registration and login with secure cookies
- 🗂️ PostgreSQL persistence using JSON columns
- ⚡ Hot-swappable route handling
- 🧠 Template helpers for common operations (currency, math, HTML escaping, etc.)

---

## 🧰 Components

- **Routes**: Declarative or programmatic, privilege-aware
- **Render Engine**: Go templates with built-in helpers
- **WebSocket Server**: Integrated via `wsrooms`
- **Database Access**: Auto-creates tables, supports dynamic access via keys
- **Authentication**: Salted SHA-256 hashes with cookie-based sessions

---

## 📦 Directory Structure

```
wsframe/
├── app.go           # Main App struct, startup logic
├── auth.go          # Register/login/logout handlers
├── db.go            # PostgreSQL logic (CRUD)
├── render.go        # Template rendering and response
├── routes.go        # Route matching and configuration
├── templatefuncs.go # Custom template functions
```

---

## 🧱 Route Configuration

A `Route` object can specify:
- A regex path match
- Optional `admin` and `authorized` variants
- Data table and key for dynamic loading
- Template to render
- Controllers to execute in frontend

Example:

```json
{
  "route": "/articles/([a-z0-9-]+)",
  "authorized": {
    "privilege": "user",
    "template": "article",
    "controllers": "articleController",
    "table": "articles",
    "key": "$1"
  }
}
```

---

## 🧪 Custom Handlers

You can register WebSocket route handlers programmatically:

```go
app.AddRoute("^/test/(.*)$", func(c *wsrooms.Conn, msg *wsrooms.Message, matches []string) {
    app.Render(c, msg, "test-view", []string{"test"}, nil)
})
```

---

## 🛠 Template Helpers

- `usd`, `add`, `subtract`, `multiply`, `divide`
- `tokey`, `fromkey`
- `sha1sum`, `css`, `unescaped`

Use them in your templates:

```gohtml
<p>{{ usd .Price }}</p>
```

# 🧪 WSFrame Example App

This is a minimal example application using [WSFrame](https://github.com/joncody/wsframe), a Go WebSocket framework for real-time routing and template rendering.

---

## 🔧 Features

- Hash-based routing (e.g. `#/test/foo`)
- Real-time template + controller updates
- Dynamic path-based data loading
- User registration & login with secure cookies

---

## 🚀 Quickstart

### 1. Clone

```bash
git clone https://github.com/joncody/wsframe.git
cd wsframe/example
```

### 2. Setup Database

Create a PostgreSQL database matching the config:

```sql
CREATE DATABASE mydatabase;
CREATE USER myuser WITH PASSWORD 'mypassword';
GRANT ALL PRIVILEGES ON DATABASE mydatabase TO myuser;
```

### 3. Run

```bash
go run main.go
```

Open in your browser:

```
http://localhost:9001
```

---

## 🗂️ Project Structure

```
.
├── main.go              # Entry point with custom handler
├── config.json          # App config
├── static/
│   ├── views/           # Go templates
│   └── js/              # Frontend client
└── wsframe/             # Framework (submodule or local copy)
```

---

## 🔒 Auth Endpoints

| Method | Path      | Description   |
|--------|-----------|---------------|
| POST   | `/register` | Register new user |
| POST   | `/login`  | Log in with alias/passhash |
| POST   | `/logout` | Logout and clear cookie |

---

## 🧪 Custom Route

Inside `main.go`:

```go
app.AddRoute("^/test/(.*)$", testHandler)

func testHandler(c *wsrooms.Conn, msg *wsrooms.Message, matches []string) {
	log.Println(matches)
	app.Render(c, msg, "index-added", []string{"index"}, nil)
}
```

Visit `http://localhost:9001/#/test/hello` to see it in action.

---

## 🌐 Frontend: `static/js/app.js`

- Listens for hash changes
- Sends `request` to WebSocket
- Receives `response` with:
  - HTML template to inject
  - JS controllers to run

---

## ✨ Example HTML Controller

```html
<a data-href="/test/hello">Go to Hello</a>
```

Triggers:

1. `location.hash = "/test/hello"`
2. WebSocket sends request
3. Server responds with rendered `index-added` view

---

## 📜 License

See the [LICENSE](./LICENSE) file for details.
