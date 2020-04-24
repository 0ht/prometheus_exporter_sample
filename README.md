# exporter-sample試行

## Dockerfileの作成

以下のDockerfileを作成する。

```Dockerfile
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
```

## Docker イメージをビルド

以下のコマンドでビルド

```sh
$ docker build --rm -f "Dockerfile" -t exporter_sample:0.1 "."
Sending build context to Docker daemon  24.08MB
Step 1/8 : FROM golang:latest
 ---> 25c4671a1478
Step 2/8 : WORKDIR /go/src/app
 ---> Using cache
 ---> 9e4fd1d0ae01
Step 3/8 : ADD ./vic_exporter.go /go/src/app
 ---> 7383e5a7abb1
Step 4/8 : RUN go get -d -v "github.com/prometheus/client_golang/prometheus"
 ---> Running in 508bf16c78dd
github.com/prometheus/client_golang (download)
github.com/beorn7/perks (download)
github.com/cespare/xxhash (download)
github.com/golang/protobuf (download)
github.com/prometheus/client_model (download)
github.com/prometheus/common (download)
github.com/matttproud/golang_protobuf_extensions (download)
github.com/prometheus/procfs (download)
get "golang.org/x/sys/unix": found meta tag get.metaImport{Prefix:"golang.org/x/sys", VCS:"git", RepoRoot:"https://go.googlesource.com/sys"} at //golang.org/x/sys/unix?go-get=1
get "golang.org/x/sys/unix": verifying non-authoritative meta tag
golang.org/x/sys (download)
Removing intermediate container 508bf16c78dd
 ---> e5e2769ca67e
Step 5/8 : RUN go get -d -v "github.com/prometheus/client_golang/prometheus/promhttp"
 ---> Running in fd33d1002544
Removing intermediate container fd33d1002544
 ---> c10fa3f4e970
Step 6/8 : RUN go get -d -v "github.com/koron/go-dproxy"
 ---> Running in 9191e894f652
github.com/koron/go-dproxy (download)
Removing intermediate container 9191e894f652
 ---> 4e4db5827271
Step 7/8 : RUN go build -o vic_exporter ./vic_exporter.go
 ---> Running in e8fb9e42afac
Removing intermediate container e8fb9e42afac
 ---> 572b9b2a9a10
Step 8/8 : CMD ["/go/src/app/exporter_sample"]
 ---> Running in bd157e8715c9
Removing intermediate container bd157e8715c9
 ---> f8a2105be509
Successfully built f8a2105be509
Successfully tagged exporter_sample:0.1
```

## Docker hubにPUSH
本来はきちんとCI/CD経由でデプロイする必要があるが、ひとまず試行ということでDocker hub経由で手動デプロイ
docker hub にログインしてPUSH

```sh
$ dcoker login
-bash: dcoker: command not found
$ docker login
Authenticating with existing credentials...
Login Succeeded
$ docker tag exporter-sample:0.1 ohtom/exporter-sample:0.1
$ docker push ohtom/exporter-sample:0.1
The push refers to repository [docker.io/ohtom/exporter-sample]
9461bc7a8b20: Pushed 
1eb010a76289: Pushed 
2ea821fe63b9: Pushed 
2bfc716e84ff: Pushed 
b8bd796cdf05: Pushed 
f212d1a5cb4b: Layer already exists 
21640a008db2: Layer already exists 
b83ca46707d6: Layer already exists 
647ae2cee1ef: Layer already exists 
6670e930ed33: Layer already exists 
c7f27a4eb870: Layer already exists 
e70dfb4c3a48: Layer already exists 
1c76bd0dc325: Layer already exists 
0.1: digest: sha256:bXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX size: 3048
```

## exporter-sample のデプロイ

以下のdeployment用、service用のマニフェストファイルを用意し、適用する。

```deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: exporter-sample
  namespace: istio-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: exporter-sample
  template:
    metadata:
      labels:
        app: exporter-sample
    spec:
      containers:
      - name: exporter-sample
        image: ohtom/exporter-sample:0.1
        imagePullPolicy: Always
        ports:
        - containerPort: 9080
```

```services.yaml
apiVersion: v1
kind: Service
metadata:
  name: exporter-sample
  namespace: istio-system
spec:
  type: ClusterIP
  selector:
    app: exporter-sample
  ports:
  - name: exporter-sample
    port: 9080
    targetPort: 9080
    protocol: TCP
```

これらのファイルをbastionサーバーにscp

```sh
$ scp -r -i ~/.ssh/id_rsa ./k8s-manifests/ admin@bastion-XXX.ap-northeast-1.elb.amazonaws.com:~/
```

bastionにログインして適用

```sh
$ kubectl apply -f deployment.yml 
deployment.apps/exporter-sample created
$ kubectl apply -f service.yml 
service/exporter-sample created
```
