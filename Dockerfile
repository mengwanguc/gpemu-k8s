FROM golang:1.10.1 as build

WORKDIR /go/src/gpemu-k8s-device-plugin

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only


COPY . .
RUN export CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' && \
    go install -ldflags="-s -w" -v gpemu-k8s-device-plugin

FROM debian:stretch-slim
COPY --from=build /go/bin/gpemu-k8s-device-plugin /bin/gpemu-k8s-device-plugin

CMD ["/bin/gpemu-k8s-device-plugin"]
