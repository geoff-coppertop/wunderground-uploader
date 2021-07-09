# Compile stage
FROM golang:1.16.5 AS build-env
ARG REPO=wunderground-uploader
ENV CGO_ENABLED 0

ADD . /$REPO

WORKDIR /$REPO
RUN make compile

# Final stage
FROM scratch
ARG REPO=wunderground-uploader
ARG OWNER=geoff-coppertop
ARG BASE_URL=https://github.com
LABEL org.opencontainers.image.source $BASE_URL/$OWNER/$REPO

COPY --from=build-env /$REPO/bin/$REPO /

# Run
CMD ["/$REPO"]