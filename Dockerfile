FROM golang:alpine as build

WORKDIR /build
RUN apk add git --no-cache
ADD go.mod ./
ADD go.sum ./
RUN go mod download -x

ADD . ./
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /bot .

FROM golang:alpine
COPY --from=build /bot /bot
