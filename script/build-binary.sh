mkdir -p ../build

cd ..

VER=$(cat VERSION)
GITHASH=$(git rev-parse --short HEAD)
BUILDTIME=`date "+%Y-%m-%dT%H:%M:%S"`
echo version is $VER
echo commit hash is $GITHASH
echo Build time is $BUILDTIME

env GOOS=darwin GOARCH=amd64 go build -v -ldflags "-X 'main.AppVer=$VER' -X 'main.BuildTime=$BUILDTIME' -X 'main.GitHash=$GITHASH'" -o _/darwin-amd64/lark *.go 
env GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.AppVer=$VER' -X 'main.BuildTime=$BUILDTIME' -X 'main.GitHash=$GITHASH'" -o _/linux-amd64/lark *.go
env GOOS=linux GOARCH=arm64 go build -v -ldflags "-X 'main.AppVer=$VER' -X 'main.BuildTime=$BUILDTIME' -X 'main.GitHash=$GITHASH'" -o _/linux-arm64/lark *.go
env GOOS=windows GOARCH=amd64 go build -v -ldflags "-X 'main.AppVer=$VER' -X 'main.BuildTime=$BUILDTIME' -X 'main.GitHash=$GITHASH'" -o _/windows-amd64/lark.exe *.go