BINARY=genjsonschema-cli
GOARCH=amd64
GOOS=linux
VERSION=unknown
LDFLAGS=-ldflags "-X main.VERSION=${VERSION}"

build:
	GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY} .

clean:
	rm -f ${BINARY}
