#! /usr/bin/env bash

# wrapper to fetch/generate/build vtproto toolchain
# for protobufs

set -x
Z=`basename $0`
PWD=`pwd`
Go=`which go`

die() {
    echo "$Z: $@" 1>&2
    exit 0
}

warn() {
    echo "$Z: $@" 1>&2
}

case $BASH_VERSION in
    4.*|5.*) ;;

    *) die "I need bash 4.x to run!"
        ;;
esac

HostOS=$($Go  env GOHOSTOS)   || die "can't get Go HostOs"
HostCPU=$($Go env GOHOSTARCH) || die "can't get Go HostCPU"
Hostbindir=$PWD/bin/$HostOS-$HostCPU

export PATH=$Hostbindir:$PATH

# build a tool that runs on the host - if needed.
hosttool() {
    local tool=$1
    local bindir=$2
    local src=$3

    local p=$(type -P $tool)
    if [ -n "$p" ]; then
        echo $p
        return 0
    fi

    p=$bindir/$tool
    if [ -x $p ]; then
        echo $p
        return 0
    fi

    local tmpdir=/tmp/$tool.$$
    mkdir $tmpdir || die "can't make $tmpdir"

    # since go1.20 - install uses env vars to decide where to put
    # build artifacts. Why are all the google tooling so bloody dev
    # hostile! WTF is wrong with command line args?!
    export GOBIN=$bindir

    # build it and stash it in the hostdir
    echo "Building tool $tool from $src .."
    (
       cd $tmpdir 
       $e $Go install  $src@latest || die "can't install $tool"
    )
    $e rm -rf  $tmpdir
    return 0
}


buildproto() {
    local pbgo=protoc-gen-go
    local vtgo=protoc-gen-go-vtproto
    local vtgo_src=github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto
    local pc
    local args="$*"

    local pgen=$(type -p protoc)
    local gogen=$(type -p $pbgo)
    local vt=$Hostbindir/$vtgo

    [ -z $pgen  ] && die "install protoc tools"
    [ -z $gogen ] && die "install protoc-gen-go"

    # now install the vtproto generator
    hosttool $vtgo $Hostbindir $vtgo_src

    for f in $args; do
        local dn=$(dirname $f)
        local bn=$(basename $f .proto)


        $e $pgen  \
            --go_out=. --plugin protoc-gen-go=$gogen \
            --go-vtproto_out=. --plugin protoc-gen-go-vtproto="$vt" \
            --go-vtproto_opt=features=marshal+unmarshal+size  \
             $f
    done
    
}


buildproto internal/proto/xstat.proto


