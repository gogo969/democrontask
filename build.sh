#! /bin/bash

PROJECT="crontask"

git checkout main
git pull origin main
git submodule init
git submodule update

GitReversion=`git rev-parse HEAD`
BuildTime=`date +'%Y.%m.%d.%H%M%S'`
BuildGoVersion=`go version`

go build -ldflags "-X main.gitReversion=${GitReversion}  -X 'main.buildTime=${BuildTime}' -X 'main.buildGoVersion=${BuildGoVersion}'" -o $PROJECT
mv $PROJECT /opt/deploy/cg/$PROJECT
cd /opt/deploy/cg/$PROJECT

git pull
git commit -am "${GitReversion}"
git push

