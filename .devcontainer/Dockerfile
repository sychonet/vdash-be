ARG GOLANG_VERSION=1.23
FROM --platform=${BUILDPLATFORM} golang:${GOLANG_VERSION}
EXPOSE 7201
#COPY .netrc /root/.netrc
#RUN go env -w GOPRIVATE=github.com/tiket/*
# gopls - go language server (https://pkg.go.dev/golang.org/x/tools/gopls)
RUN go install -v golang.org/x/tools/gopls@latest
# dlv -  debugger for golang (https://github.com/go-delve/delve)
RUN go install -v github.com/go-delve/delve/cmd/dlv@latest
# staticcheck - linter for golang (https://staticcheck.io/)
RUN go install -v honnef.co/go/tools/cmd/staticcheck@latest
