FROM golang:1.24-alpine AS build

WORKDIR /app

COPY container_src/go.mod ./
RUN go mod download;

COPY container_src/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /server

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /server /server
COPY container_src/prompts/ /prompts/
EXPOSE 8080

CMD ["/server"]
