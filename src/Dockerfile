FROM golang:alpine AS builder  
RUN apk add --no-cache git
WORKDIR /tmp/go-app  
COPY go.mod .

RUN go mod download

COPY . .

RUN go mod tidy && go mod vendor && CGO_ENABLED=0 GOOS=linux go build -o main .  
  
FROM scratch  
WORKDIR /  
COPY --from=builder /tmp/go-app/main main  
ENTRYPOINT ["/main"]