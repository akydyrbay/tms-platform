FROM golang:1.26.3-alpine AS backend_builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go


FROM alpine:3.20.1 AS backend

WORKDIR /app
COPY --from=backend_builder /app/main .

EXPOSE 8080
CMD ["./main"]


FROM node:20-alpine AS frontend

WORKDIR /frontend

COPY frontend/package*.json ./
RUN npm install

COPY frontend/ .

EXPOSE 5173

CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]