# NorthPost Service

[![CI](https://github.com/NorthPostApp/service/actions/workflows/ci.yml/badge.svg)](https://github.com/NorthPostApp/service/actions/workflows/ci.yml)

Backend service for north post app

docker build -t north-post-dev:1.0.0 .

docker run -p 8080:8080 north-post-dev

docker-compose build north-post-admin

docker-compose up north-post-admin