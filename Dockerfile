FROM golang:1.26 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/skillforge-api ./cmd/skillforge-api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/skillforge-api /app/skillforge-api

ENV SKILLFORGE_LISTEN_ADDR=:8080
ENV SKILLFORGE_REPO_ROOT=/data/skills-repo
EXPOSE 8080

ENTRYPOINT ["/app/skillforge-api"]
