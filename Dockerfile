FROM golang:1.10.3
ENV PATH=/usr/local/Cellar/go/1.10.3/libexec/bin:$PATH
WORKDIR $GOPATH/src/BlockChainTest
ADD . $GOPATH/src/BlockChainTest
#RUN go build main.go
EXPOSE 55555
CMD go run main.go
#ENTRYPOINT  ["./main"]
