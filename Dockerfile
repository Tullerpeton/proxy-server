FROM golang:1.15 AS builder

WORKDIR /build

COPY . .

RUN go build ./cmd/proxy-server/main.go

FROM ubuntu:20.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install postgresql-12 -y

USER postgres

COPY ./scripts/sql/init_db.sql .

RUN service postgresql start && \
        psql -c "CREATE USER proxy_user WITH superuser login password 'proxy_user';" && \
        psql -c "ALTER ROLE proxy_user WITH PASSWORD 'proxy_user';" && \
        createdb -O proxy_user proxy_db && \
        psql -f init_db.sql -d proxy_db && \
        service postgresql stop

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /proxy
COPY --from=builder /build/main .

COPY . .

EXPOSE 8080
EXPOSE 8000
EXPOSE 5432

ENV PROXY_PORT=8080
ENV REPEATER_PORT=8000
ENV DB_USER=user
ENV DB_NAME=Requests

CMD service postgresql start && ./main