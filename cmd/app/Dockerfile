FROM docker.io/library/golang:1.23 as build

WORKDIR /src

COPY ./.. .

RUN go mod tidy
RUN go build -o app cmd/app/main.go


FROM docker.io/library/golang:1.23

WORKDIR /src

COPY --from=build /src/app app

CMD ["./app"]
