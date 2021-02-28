FROM golang:buster as builder
ENV APP_USER app
ENV APP_HOME /go/src/vegabot
RUN groupadd $APP_USER && useradd -m -g $APP_USER -l $APP_USER
RUN mkdir -p $APP_HOME && chown -R $APP_USER:$APP_USER $APP_HOME
WORKDIR $APP_HOME
#USER $APP_USER
COPY src/ .
RUN ls
RUN go mod download
RUN go mod verify
RUN go build -o vegabot github.com/baldator/vega-alerts

FROM alpine:3.12.4
ENV APP_USER app
ENV APP_HOME /go/src/vegabot
RUN apk add --no-cache libc6-compat
RUN addgroup $APP_USER && adduser -S $APP_USER -G $APP_USER
RUN mkdir -p $APP_HOME
WORKDIR $APP_HOME
COPY src/config.yaml $APP_HOME/config/config.yaml
COPY --chown=0:0 --from=builder $APP_HOME/vegabot $APP_HOME
USER $APP_USER
CMD ["./vegabot"]