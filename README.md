Project structure

```plain
myapp/
├── cmd/
│   └── server/
│       └── main.go          // entrypoint: wire dependencies, start server
├── internal/
│   ├── config/
│   │   └── config.go        // load env vars / config files
│   ├── handler/             // HTTP handlers (gin.Context lives here only)
│   │   ├── user.go
│   │   └── health.go
│   ├── service/             // business logic, framework-agnostic
│   │   └── user.go
│   ├── repository/          // data access (DB, cache, external APIs)
│   │   └── user.go
│   ├── model/               // domain types / structs
│   │   └── user.go
│   ├── middleware/          // auth, logging, recovery, CORS
│   │   └── auth.go
│   └── router/
│       └── router.go        // route definitions, group setup
├── pkg/                     // optional: reusable code safe to import elsewhere
├── go.mod
└── go.sum
```
