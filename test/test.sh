#!/bin/bash
#
#  test.sh
#

CURRENT=$(cd $(dirname $0);pwd)

export PATH=$CURRENT/../bin:$PATH

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

#
#  テスト実施
#
function f-test-convert() {
    local input_file=$1
    shift
    local other_opt="$@"
    local base_file_name=${input_file%%.yaml}
    local override_file=${base_file_name}-override.yaml
    local output_file=out1/${base_file_name}-out.yaml
    local answer_file=ans1/${base_file_name}-ans.yaml
    local output_file2=out2/${base_file_name}-out2.yaml
    local answer_file2=ans2/${base_file_name}-ans2.yaml

    mkdir -p out1 ans1 out2 ans2

    # pass 1
    if [ -f $override_file ]; then
        f-test-success yamlsort -i $input_file -o $output_file --override-file $override_file ${other_opt}
    else
        f-test-success yamlsort -i $input_file -o $output_file ${other_opt}
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
        f-test-success yamlsort -i $output_file -o $output_file2 --override-file $override_file ${other_opt}
    else
        f-test-success yamlsort -i $output_file -o $output_file2 ${other_opt}
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


#
#  テスト実施(json output)
#
function f-test-convert-json() {
    local input_file=$1
    shift
    local other_opt="$@"
    local base_file_name=${input_file%%.yaml}
    local override_file=${base_file_name}-override.yaml
    local output_file=out1/${base_file_name}-out.json
    local answer_file=ans1/${base_file_name}-ans.json
    local output_file2=out2/${base_file_name}-out2.yaml
    local answer_file2=ans2/${base_file_name}-ans2.yaml

    mkdir -p out1 ans1 out2 ans2

    # pass 1
    if [ -f $override_file ]; then
        f-test-success yamlsort -i $input_file -o $output_file --override-file $override_file ${other_opt} --jsonoutput
    else
        f-test-success yamlsort -i $input_file -o $output_file ${other_opt} --jsonoutput
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
        f-test-success yamlsort -i $output_file --jsoninput -o $output_file2 --override-file $override_file ${other_opt}
    else
        f-test-success yamlsort -i $output_file --jsoninput -o $output_file2 ${other_opt}
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

# basic テスト。 map の key 項目ソートで、 name が a よりも前にくること。
f-log "convert 1 : map key name sort. name as top key when sorted output."
f-test-convert  sample1.yaml

f-log "convert 2 : pvc, service, deployment multi file yaml test."
f-test-convert  sample2.yaml

f-log "convert 3 : ingress yaml test."
f-test-convert  sample3.yaml

f-log "convert 4 : override test. same key is override. map[name=key] data is override."
f-test-convert  sample4.yaml

f-log "convert 5 : override test. append list item. "
f-test-convert  sample5.yaml

f-log "convert 6 : string quote test."
f-test-convert  sample6.yaml

f-log "convert 7 : override test. same key is override."
f-test-convert  sample7.yaml

f-log "convert 8 : override test. same key is override. map[name=key] data is override."
f-test-convert  sample8.yaml

f-log "convert 9 : string quote test. boolean key word true/false, yes/no, on/off"
f-test-convert  sample9.yaml

f-log "convert 10 : override test. map[name=key] data is override."
f-test-convert  sample10.yaml

f-log "convert 11 : check --skip-key optoin"
f-test-convert  sample11.yaml --skip-key  spec.template.spec.containers[name=kjwikigdocker-container].env[name=abc]

f-log "convert 12 : test zero length array, zero length map"
f-test-convert  sample12.yaml

f-log "convert 13 : check --select-key option"
f-test-convert  sample13.yaml --select-key spec.template.spec.containers[name=kjwikigdocker-container].env[name=abc]  --select-key spec.template.spec.containers[name=kjwikigdocker-container].name

f-log "convert 14 : check zero-length string"
f-test-convert  sample14.yaml

f-log "convert 15 : check yaml here document and jsonoutput"
f-test-convert-json  sample15.yaml

f-log "TEST_SUCCESS_COUNT  $TEST_SUCCESS_COUNT  "
f-log "TEST_FAILURE_COUNT  $TEST_FAILURE_COUNT  "
