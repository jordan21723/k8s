FROM golang:1.15-alpine3.12 AS builder

ARG VERSION

RUN apk add gcc musl-dev
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct go build -o dist/server -a -ldflags "-w -s -X k8s-installer/pkg/server/version.Version=${VERSION}" ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOPROXY=https://goproxy.cn,direct go run ./tools/doc-gen/main.go --output=dist/swagger.json

FROM alpine:3.12

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
  && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=builder /app/dist/server /app/server
COPY --from=builder /app/dist/swagger.json /app/swagger.json

EXPOSE 8099 9889 8079

CMD ["/app/server"]
