FROM golang:1.21.3-bookworm AS builder
RUN apt-get update && apt-get upgrade -y

WORKDIR /app

# get dependencies first separately
COPY go.* .
RUN go mod download

# the rest of the source code needed for building
COPY . .

RUN go build


FROM debian:bookworm-slim AS deployment
RUN apt-get update && apt-get upgrade -y
RUN apt-get install cron -y

COPY crontab.txt .
COPY --from=builder /app/forecast-notification /app/forecast-notification

# create log file to be able to run tail
RUN touch /app/cron.log

# install crontab
ENV TZ="Europe/Prague"
ENV CRON_TZ="Europe/Prague"
RUN crontab crontab.txt

# copy env variables somewhere, where cron can access them
# https://stackoverflow.com/questions/65884276
CMD printenv > /etc/environment && cron && tail -f /app/cron.log
