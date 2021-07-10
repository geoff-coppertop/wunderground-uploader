# http://www.inanzzz.com/index.php/post/1sfg/multi-stage-docker-build-for-a-golang-application-with-and-without-vendor-directory

# Compile stage
FROM golang:1.16.5 AS build-env
ENV CGO_ENABLED 0

WORKDIR /wunderground-uploader

ADD . ./

RUN make build

# Final stage
FROM scratch
LABEL org.opencontainers.image.source https://github.com/geoff-coppertop/wunderground-uploader

COPY --from=build-env /wunderground-uploader/bin/wunderground-uploader /

# Run
CMD ["/wunderground-uploader"]