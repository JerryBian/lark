FROM golang:buster AS build

WORKDIR /app

COPY . ./
RUN go mod download

RUN go build -o /lark

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /lark /lark

ENV PORT=9000
ENV GIN_MODE=release
ENV DB_LOCATION=/db
EXPOSE 9000

USER nonroot:nonroot

ENTRYPOINT ["/lark"]