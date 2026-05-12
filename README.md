# CRUD Service for the NorthPost

[![CI](https://github.com/NorthPostApp/service/actions/workflows/ci.yml/badge.svg)](https://github.com/NorthPostApp/service/actions/workflows/ci.yml)

## Gin Swagger Documentation

1. Install swagger globally
```
go install github.com/swaggo/swag/cmd/swag@latest
```
2. Run command from the root dit to generate swagger doc
```
swag init -g cmd/api/main.go --output docs
```
Once running, visit `http://localhost:<PORT>/swagger/index.html` to see the UI.

3. Add the following step to the CI/CD workflow:
```
- name: Generate Swagger docs
  run: |
    go install github.com/swaggo/swag/cmd/swag@latest
    swag init -g cmd/api/main.go --output docs
```

**Troubleshoot**
- `zsh: command not found: swag`: run this command `echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc && source ~/.zshrc`
