# The base image to install OS dependencies
FROM golang:1.14.1-alpine AS base

RUN apk update && \
  apk upgrade && \
  apk add autoconf automake bash ca-certificates git gcc g++ libc6-compat libjpeg-turbo-dev \
  libpng-dev make nodejs nodejs-npm python upx vips && \
  rm -rf /var/cache/*

# The intermediate image to build the binary
FROM base AS builder

WORKDIR /home/{{.Project.Name}}
COPY . .

RUN go mod download
RUN cd web && npm install
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go run . build

# The final image to run on production
FROM alpine:3.10.3

ENV APP_HOME=/home/{{.Project.Name}}
ENV GROUP_NAME={{.Project.Name}}
ENV USER_NAME={{.Project.Name}}
HEALTHCHECK CMD curl -f http://localhost:3000/health_check || exit 1
WORKDIR ${APP_HOME}

COPY --from=builder /home/{{.Project.Name}}/{{.Project.Name}} ${APP_HOME}

RUN apk update && apk upgrade && \
  apk add --no-cache chromium && \
  rm -rf /var/cache/apk/* /var/lib/apt/lists/* /var/cache/apk/* /usr/share/man /tmp/*
RUN addgroup -S ${GROUP_NAME} && adduser -S ${GROUP_NAME} -G ${USER_NAME} && \
  chown -R ${GROUP_NAME}:${USER_NAME} ${APP_HOME}

USER ${USER_NAME}
CMD ["./{{.Project.Name}}", "serve"]
