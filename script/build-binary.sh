mkdir -p ../build

env GOOS=darwin GOARCH=amd64 go build -v -o ../_/darwin-amd64/lark ../*.go
env GOOS=linux GOARCH=amd64 go build -v -o ../_/linux-amd64/lark ../*.go
env GOOS=linux GOARCH=arm64 go build -v -o ../_/linux-arm64/lark ../*.go
env GOOS=windows GOARCH=amd64 go build -v -o ../_/windows-amd64/lark.exe ../*.go