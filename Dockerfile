# Compile stage
FROM golang:1.16.5 AS build-env
ENV CGO_ENABLED 0

ADD . /wunderground-uploader

WORKDIR /wunderground-uploader
RUN make compile

# Final stage
FROM scratch

COPY --from=build-env /wunderground-uploader/bin/wunderground-uploader /

# Run
CMD ["/wundergound-uploader"]