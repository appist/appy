# The base image to install OS dependencies
FROM golang:1.15.2-alpine AS base

RUN apk update && \
  apk upgrade && \
  apk add autoconf automake bash ca-certificates git gcc g++ libc6-compat libjpeg-turbo-dev \
  libpng-dev make nodejs nodejs-npm python upx vips && \
  rm -rf /var/cache/*

# The intermediate image to build the binary
FROM base AS builder

WORKDIR /home/{{.projectName}}
COPY . .

RUN go mod download
RUN npm install
RUN CGO_ENABLED=1 go run . build --static --platform=linux/amd64

# The final image to run on production
FROM alpine:3.12.0

ENV APP_HOME=/home/{{.projectName}}
ENV GROUP_NAME={{.projectName}}
ENV USER_NAME={{.projectName}}
HEALTHCHECK CMD curl -f http://localhost:3000/health_check || exit 1
WORKDIR ${APP_HOME}

COPY --from=builder /home/{{.projectName}}/{{.projectName}} ${APP_HOME}

RUN apk update && apk upgrade && \
  apk add --no-cache chromium && \
  rm -rf /var/cache/apk/* /var/lib/apt/lists/* /var/cache/apk/* /usr/share/man /tmp/*
RUN addgroup -S ${GROUP_NAME} && adduser -S ${GROUP_NAME} -G ${USER_NAME} && \
  chown -R ${GROUP_NAME}:${USER_NAME} ${APP_HOME}

USER ${USER_NAME}
CMD ["./{{.projectName}}", "serve"]
