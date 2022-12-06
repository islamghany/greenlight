# Microservices project for movies.

This project contains a microservices application written mostly in Go, which will be updated periodically as I learn more about backend engineering, including scalability, security, performance, and best practices in modern system design.

## Architecture overview

![Greenlight Architecture](./statics/greenlight-arch.jpg)

[Greenlight client application(still not finished)](https://github.com/islamghany/greenlight-client)

## Tech Stack

<img alt="GoLang" src="https://img.shields.io/badge/-Golang-007D9C?style=for-the-badge&logo=go&logoColor=white" />
<img alt="postgres" src="https://img.shields.io/badge/-postgres-430098?style=for-the-badge&logo=postgresql&logoColor=white" />
<img alt="Node.js" src="https://img.shields.io/badge/-Node.JS-00A95C?style=for-the-badge&logo=node.js&logoColor=white" />
<img alt="MongoDB" src="https://img.shields.io/badge/-MongoDB-47A248?style=for-the-badge&logo=mongodb&logoColor=white" />
<img alt="redis" src="https://img.shields.io/badge/-redis-d63835?style=for-the-badge&logo=redis&logoColor=white" />
<img alt="nginx" src="https://img.shields.io/badge/-Nginx-009639?style=for-the-badge&logo=nginx&logoColor=white" />
 <img alt="rabbitmq" src="https://img.shields.io/badge/-RabbitMQ-FB693F?style=for-the-badge&logo=rabbitmq&logoColor=white" />
 <img alt="docker" src="https://img.shields.io/badge/-docker-1572B6?style=for-the-badge&logo=docker&logoColor=white" />
 <img alt="grpc" src="https://img.shields.io/badge/-gRPC-2B545C?style=for-the-badge&logo=rpc&logoColor=white" />
 <img alt="elastic" src="https://img.shields.io/badge/-ElasticSearch-FDC63C?style=for-the-badge&logo=elasticsearch&logoColor=white" />
 <img alt="Ts" src="https://img.shields.io/badge/-typescript-3178C6?style=for-the-badge&logo=typescript&logoColor=white" />
 <img alt="PASETO" src="https://img.shields.io/badge/-PASETO-1a1a1a?style=for-the-badge&logo=paseto&logoColor=white" />
<img alt="protobuf" src="https://img.shields.io/badge/-protobuf-D44235?style=for-the-badge&logo=protobuf&logoColor=white" />
<img alt="swagger" src="https://img.shields.io/badge/-Swagger-D44235?style=for-the-badge&logo=swagger&logoColor=white" />

## Installation

The installation is very simple the only required thing is to have docker and docker-compose installed on your machine.

First clone the project

```bash
git clone https://github.com/islamghany/greenlight.git
```

then go the project folder in inside greenlight project.

```bash
cd ./greenlight/project
```

then run the docker compose up to generate containers.

```bash
docker compose up
```

to stop the containers hit _Ctrl+C_

to Stop containers and removes containers, networks, volumes, and images created by docker compose up command.

```bash
docker compose down
```

if you made a change in the code and you want to see just rebuild the images with

```bash
docker compose up --build
```

## Usage

After installation, you may want to interact with the project.
So, you can open the swagger API and play with the project APIs through this URL

```bash
http://localhost:3050/v1/swagger/
```

![Swagger](./statics/swagger.png)

Also if you registered as a user you may want to check the mail through the mail service dashboard, since an activation token
is sent to your registered mail.

you can open the email-testing dashboard through the link

```bash
http://localhost:8025/
```

![Swagger](./statics/mailhog.png)
