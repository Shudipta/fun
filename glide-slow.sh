#!/bin/bash

# You can execute me through Glide by doing the following:
# - Execute `glide slow`
# - ???
# - Profit

pushd $GOPATH/src/fun

glide up -v
glide vc --only-code --no-tests

popd
