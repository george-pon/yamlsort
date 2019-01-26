#!/bin/bash
#
#  test.sh
#

source ./test-common.sh

function f-log() {
    echo "==== $@"
}

function f-cmd-run() {
    echo "$@"
    "$@"
}

function f-test-success() {
    echo "$@"
    if "$@" ; then
        echo "SUCCESS"
        TEST_SUCCESS_COUNT=$(( $TEST_SUCCESS_COUNT + 1 ))
        return 0
    else
        echo "FAILURE"
        TEST_FAILURE_COUNT=$(( $TEST_FAILURE_COUNT + 1 ))
        return 1
    fi
}

function f-test-failure() {
    echo "$@"
    if "$@" ; then
        echo "FAILURE"
        TEST_FAILURE_COUNT=$(( $TEST_FAILURE_COUNT + 1 ))
        return 1
    else
        echo "SUCCESS"
        TEST_SUCCESS_COUNT=$(( $TEST_SUCCESS_COUNT + 1 ))
        return 0
    fi
}

function f-test-convert() {
    local input_file=$1
    local base_file_name=${input_file%%.yaml}
    local override_file=${base_file_name}-override.yaml
    local output_file=${base_file_name}-out.yaml
    local answer_file=${base_file_name}-ans.yaml
    local output_file2=${base_file_name}-out2.yaml
    local answer_file2=${base_file_name}-ans2.yaml
    if [ -f $override_file ]; then
        f-test-success yamlsort -i $input_file -o $output_file --override-file $override_file
    else
        f-test-success yamlsort -i $input_file -o $output_file
    fi
    if [ -f $answer_file ]; then
        if diff -u $answer_file $output_file ; then
            echo "diff SUCCESS"
            TEST_SUCCESS_COUNT=$(( $TEST_SUCCESS_COUNT + 1 ))
        else
            echo "diff $answer_file $output_file FAILURE"
            TEST_FAILURE_COUNT=$(( $TEST_FAILURE_COUNT + 1 ))
        fi
    else
        cp $output_file $answer_file
    fi
    # pass 2
    if [ -f $override_file ]; then
        f-test-success yamlsort -i $output_file -o $output_file2 --override-file $override_file
    else
        f-test-success yamlsort -i $output_file -o $output_file2
    fi
    if [ -f $answer_file2 ]; then
        if diff -u $answer_file2 $output_file2 ; then
            echo "diff SUCCESS"
            TEST_SUCCESS_COUNT=$(( $TEST_SUCCESS_COUNT + 1 ))
        else
            echo "diff $answer_file2 $output_file2 FAILURE"
            TEST_FAILURE_COUNT=$(( $TEST_FAILURE_COUNT + 1 ))
        fi
    else
        cp $output_file2 $answer_file2
    fi
}

TEST_SUCCESS_COUNT=0
TEST_FAILURE_COUNT=0

f-log "version"
f-test-success yamlsort version

f-log "convert "
f-test-convert  sample.yaml

f-log "convert 1"
f-test-convert  sample1.yaml

f-log "convert 2"
f-test-convert  sample2.yaml

f-log "convert 3"
f-test-convert  sample3.yaml

f-log "convert 4"
f-test-convert  sample4.yaml

f-log "convert 5"
f-test-convert  sample5.yaml

f-log "convert 6"
f-test-convert  sample6.yaml

f-log "convert 7"
f-test-convert  sample7.yaml

f-log "convert 8"
f-test-convert  sample8.yaml

f-log "convert 9"
f-test-convert  sample9.yaml

f-log "TEST_SUCCESS_COUNT  $TEST_SUCCESS_COUNT  "
f-log "TEST_FAILURE_COUNT  $TEST_FAILURE_COUNT  "
