FROM golang:1.26 AS builder

WORKDIR /src

ARG VERSION=dev
ARG COMMIT=unknown

COPY api/go.mod api/go.sum ./api/
RUN cd api && go mod download

COPY . .

RUN CGO_ENABLED=0 go build \
	-ldflags "-X main.version=${VERSION} -X main.hash=${COMMIT}" \
	-o ./dist/website-api ./api

FROM alpine:latest

WORKDIR /srv/

COPY --from=builder /src/dist/website-api ./website-api

CMD [ "./website-api" ]
