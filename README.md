# Project structure

Some of the layers here are not implemented in the actual project but remain as reference for the future changes.

```plain
src/
├── cmd/
│   └── server/
│       └── main.go                    // entrypoint: wire dependencies, start server
├── internal/
│   ├── configgen/
│   │   ├── templates/
│   │   │   └── ...                    // templates for config generation
│   │   ├── registry.go                // generation strategies registry
│   │   └── ...protocols.go            // various strategies of generation for different protocols (e.g. awg.go, wg.go, etc)
│   ├── executor/
│   │   └── docker.go                  // Bash executor for Docker containers
│   ├── repository/                    // data access (DB, cache, external APIs)
│   │   └── user.go
│   ├── sshclient/
│   │   ├── client.go                  // SSH client for command execution in non-container environments
│   │   └── command.go                 // base for SSH command
│   ├── tgbot/                         // Telegram bot module
│   │   ├── handlers/                  // various bot handlers
│   │   │   ├── generate-new-config.go
│   │   │   ├── start.go
│   │   │   ├── manage-keys.go
│   │   │   └── handlers.go            // base for bot handler
│   │   ├── middleware/                // bot's middleware
│   │   │   └── middleware.go
│   │   └── bot.go                    // bot's orchestration
│   ├── service/                       // business logic, framework-agnostic
│   │   └── user.go
│   ├── model/                         // domain types / structs
│   │   └── user.go
│   ├── config/                        // NOT PRESENT
│   │   └── config.go                  // NOT PRESENT load env vars / config files
│   ├── handler/                       // NOT PRESENT HTTP handlers (gin.Context lives here only)
│   │   ├── user.go
│   │   └── health.go
│   ├── middleware/                    // NOT PRESENT auth, logging, recovery, CORS
│   │   └── auth.go
│   └── router/
│       └── router.go                  // NOT PRESENT route definitions, group setup
├── pkg/                               // NOT PRESENT optional: reusable code safe to import elsewhere
├── go.mod
└── go.sum
```
