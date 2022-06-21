FROM golang:1.14-alpine AS builder

ARG VERSION

RUN apk update && mkdir /app
RUN apk add gcc musl-dev
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=1 GOOS=linux GOPROXY=https://goproxy.cn,direct go build -o dist/client -a -ldflags "-w -s -X k8s-installer/pkg/server/version.Version=${VERSION}" ./cmd/client

FROM alpine:3.10

ARG REPO_URL
ARG BRANCH
ARG COMMIT_REF
LABEL repo-url=${REPO_URL}
LABEL branch=${BRANCH}
LABEL commit-ref=${COMMIT_REF}


RUN apk update \
  && apk upgrade \
  && apk add --no-cache \
  ca-certificates tzdata \
  && update-ca-certificates 2>/dev/null || true \
  && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
  && mkdir /app
WORKDIR /app
COPY --from=builder /app/dist/client /app/client

EXPOSE 8080

CMD ["/app/client"]
