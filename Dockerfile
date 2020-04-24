# ベースとなるDockerイメージ指定
FROM golang:latest

# コンテナログイン時のディレクトリ指定
WORKDIR /go/src/app
# ホストのファイルをコンテナの作業ディレクトリに移行
ADD ./exporter_sample.go /go/src/app

RUN go get -d -v "github.com/prometheus/client_golang/prometheus"
RUN go get -d -v "github.com/prometheus/client_golang/prometheus/promhttp"
RUN go get -d -v "github.com/koron/go-dproxy"
RUN go build -o exporter_sample ./exporter_sample.go

CMD ["/go/src/app/exporter_sample"]
