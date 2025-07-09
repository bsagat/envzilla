# 🐲 EnvZilla

**EnvZilla** is a minimalist Go package for loading environment variables from config files. Perfect for development, testing, and deployment environments.

---

## 🔧 Features

- 📄 Reads configuration files
- 🧼 Ignores comments (`#`) and empty lines
- 🌱 Sets variables into the global `os.Environ`
- ✅ Support for `string`, `int`, `float`, `bool` types when working with configs via tags
- 🪐 Autoloading of variables when importing `autoload`
---

## 📦 Installation

Install the package:

```bash
go get github.com/bsagat/envzilla

For automatic variable loading::

```bash
go get github.com/bsagat/envzilla/autoload
```
---

## 🚀 Quick Start
Create `.env`:
```
DATABASE_URL=postgres://localhost:5432/mydb
DEBUG=true
PORT=8080
```
Load in the application:
```go
package main

import (
    "github.com/bsagat/envzilla"
)

func main(){
    if err := envzilla.Load(); err != nil{
        // Error handling...
    }
    // now DATABASE_URL, DEBUG and PORT are available via os.Getenv
}
```

## 🪄 Load to structure
```go
package main

import (
	"github. com/bsagat/envzilla"
)

type Config struct {
	DatabaseURL string `env: "DATABASE_URL" default: "postgres://localhost:5432/defaultdb"`
	Debug       bool   `env: "DEBUG" default: "false"`
	Port        int    `env: "PORT" default: "3000"`
}

func main() {
	var cfg Config
	if err := envzilla.Load(); err != nil {
        // Error handling...
	}
	if err := envzilla.Parse(&cfg); err != nil {
        // Error handling...
	}
	log.Println(cfg.DatabaseURL, cfg.Debug, cfg.Port)
}
```
## 🧪 Testing

Run all tests:
```
go test ./...
```
