#!/bin/bash -e

cwd=`pwd`

if [[ $1 = "-loc" ]]; then
    find . -name '*.go' | xargs wc -l | sort -n
    exit
fi

#-----------------
# install on linux
#-----------------
ACTOR_HOME=/sgn/app/actor
if [[ $1 = "-install" ]]; then
    mkdir -p $ACTOR_HOME/bin $ACTOR_HOME/var $ACTOR_HOME/etc
    cp -f daemon/actord/actord $ACTOR_HOME/bin/
    cp -f etc/actord.cf.sample $ACTOR_HOME/etc/actord.cf
    cp -f etc/actord /etc/init.d/actord
    echo 'Done'
    exit
fi

VER=0.1.5stable
ID=$(git rev-parse HEAD | cut -c1-7)

cd daemon/actord

if [[ $1 = "-linux" ]]; then
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/nicholaskh/golib/server.VERSION $VER -X github.com/nicholaskh/golib/server.BuildID $ID -w"
    exit
else
    go build -ldflags "-X github.com/nicholaskh/golib/server.VERSION $VER -X github.com/nicholaskh/golib/server.BuildID $ID -w"
    #go build -race -v -ldflags "-X github.com/nicholaskh/golib/server.BuildID $ID -w"
fi

#---------
# show ver
#---------
cd $cwd
./daemon/actord/actord -version

