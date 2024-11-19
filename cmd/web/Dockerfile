FROM golang:1.23

WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o web ./cmd/web

EXPOSE 54321
CMD ["./web"]