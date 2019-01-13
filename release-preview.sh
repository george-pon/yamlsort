#!/bin/bash
#
# release preview
#

# vi  README.md  CHANGELOG.md

# build
source build-cmd.sh

# run test
pushd test
bash test.sh
popd

