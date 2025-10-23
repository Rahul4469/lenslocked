# Tailwind v4 builder stage
FROM node:latest AS tailwind-builder
WORKDIR /tailwind
COPY ./templates /templates
COPY ./tailwind/tailwind.config.js /src/tailwind.config.js
COPY ./tailwind/styles.css /src/styles.css
RUN npm install tailwindcss@3
RUN npx tailwindcss -c /src/tailwind.config.js -i /src/styles.css -o /styles.css


FROM golang AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -v -o ./server ./cmd/server/

FROM ubuntu
WORKDIR /app
COPY ./assets ./assets
COPY .env .env
COPY --from=builder /app/server ./server
COPY --from=tailwind-builder /styles.css /app/assets/styles.css
CMD ./server

