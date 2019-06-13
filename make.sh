rm -rf supermentor
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go build
