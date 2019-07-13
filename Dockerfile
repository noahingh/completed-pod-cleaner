FROM golang:1.11 AS builder

RUN  mkdir /app
WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a -installsuffix cgo  -o /go/bin/completed-pod-cleaner ./cmd


FROM scratch
COPY --from=builder /go/bin/completed-pod-cleaner /go/bin/completed-pod-cleaner
ENTRYPOINT [ "/go/bin/completed-pod-cleaner" ]