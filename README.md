# Crowdsay

Plataforma de Enquetes e Votação em Go (Gin)

## Descrição
O Crowdsay é uma API RESTful para criação de enquetes, votação única por IP, resultados em tempo real e estatísticas percentuais.

## Endpoints

### Criar enquete
`POST /polls/create`
Body:
```json
{
	"question": "Qual sua linguagem favorita?",
	"options": ["Go", "Python", "JavaScript"]
}
```
Response:
```json
{
	"message": "poll created",
	"id": 1,
	"next": ["GET /polls/1", "POST /polls/1/vote"]
}
```

### Votar em enquete
`POST /polls/{id}/vote`
Body:
```json
{
	"option": "Go"
}
```
Headers (para simular IP):
`X-Forwarded-For: 1.2.3.4`
Response:
```json
{
	"message": "vote registered",
	"next": ["GET /polls/1"]
}
```

### Consultar enquete
`GET /polls/{id}`
Response:
```json
{
	"poll": {
		"id": 1,
		"question": "Qual sua linguagem favorita?",
		"options": ["Go", "Python", "JavaScript"],
		"votes": {"1.2.3.4:1": 1},
		"results": {"Go": 1}
	},
	"next": ["POST /polls/1/vote"]
}
```

### Estatísticas (porcentagem)
`GET /polls/{id}/stats`
Response:
```json
{
	"poll_id": 1,
	"question": "Qual sua linguagem favorita?",
	"options": ["Go", "Python", "JavaScript"],
	"votes": {"Go": 1},
	"percent": {"Go": 100, "Python": 0, "JavaScript": 0},
	"total_votes": 1,
	"next": ["GET /polls/1", "POST /polls/1/vote"]
}
```

### Listar enquetes
`GET /polls/`
Response:
```json
{
	"polls": [ ... ],
	"next": ["POST /polls/create"]
}
```

## Como rodar localmente

1. Instale Go 1.21+
2. Clone o repositório
3. Instale dependências:
	 ```bash
	 go mod tidy
	 ```
4. Rode o servidor:
	 ```bash
	 go run cmd/main.go
	 ```
5. Acesse http://localhost:8080

## Como rodar com Docker

1. Certifique-se de ter Docker instalado
2. Rode:
	 ```bash
	 docker-compose up --build
	 ```
3. A API estará disponível em http://localhost:8080

## Testes

Para rodar os testes:
```bash
go test ./internal/poll/...
```

---
Projeto para portfólio por Adriano JTF
# crowdsay
Plataforma de Enquetes e Votação
