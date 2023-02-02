FROM docker.ifeng.com/newsclient-nine/golang:go1.19-v1.1 AS build-env
ENV SRC_DIR /go/src
ENV BIN_DIR /go/bin

ADD . $SRC_DIR
RUN export GO111MODULE=on \
 && export GOPROXY=https://goproxy.io \
 && export GOARCH=amd64 GOOS=linux CGO_ENABLED=0;
RUN cd $SRC_DIR/cmd \
    && go build -v -o $BIN_DIR/httpchk httpchk.go \
    && go build -v -o $BIN_DIR/dfconsume dfconsumeMain.go \
    && go build -v -o  $BIN_DIR/statconsume statconsumeMain.go;

FROM docker.ifeng.com/newsclient-nine/golang:go1.19-v1.1
WORKDIR /app

COPY --from=build-env /go/bin/* /app/
# Copy conf files and Code files
COPY etc/supervisord.conf /etc/supervisor/supervisord.conf
EXPOSE 80
EXPOSE 8081
ENTRYPOINT ["supervisord", "-c", "/etc/supervisor/supervisord.conf"]