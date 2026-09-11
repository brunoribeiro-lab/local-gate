ARG GO_VERSION=1.22

FROM node:22-alpine AS readme-check

WORKDIR /src

RUN npm install --global markdownlint-cli2@0.18.1

COPY README.md ./

RUN markdownlint-cli2 README.md

FROM golang:${GO_VERSION}-alpine AS dependencies

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod ./

RUN go mod download

FROM dependencies AS test

COPY . .

RUN go test ./...

FROM dependencies AS builder

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w \
      -X main.version=${VERSION} \
      -X main.commit=${COMMIT} \
      -X main.buildDate=${BUILD_DATE}" \
    -o /out/localgate \
    ./cmd/localgate

FROM scratch AS binary

COPY --from=builder /out/localgate /localgate

FROM dependencies AS local-build

WORKDIR /workspace

CMD ["go", "build", "-o", "localgate", "./cmd/localgate"]

FROM alpine:3.22 AS runtime

RUN apk add --no-cache ca-certificates nginx

COPY --from=builder /out/localgate /usr/local/bin/localgate

RUN chmod +x /usr/local/bin/localgate \
    && mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled

ENTRYPOINT ["localgate"]
CMD ["help"]
