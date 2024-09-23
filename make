#!/bin/sh

DEST=tvhost_v1.004
BIN=tvhost

go build -ldflags "-linkmode external -extldflags '-static'" -o tvhost
mkdir ./$DEST
mv ./$BIN ./$DEST/$BIN
tar -cvzf $DEST.tar.gz $DEST
rm -rf ./$DEST