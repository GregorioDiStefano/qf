#!/bin/env bats
BUCKET='qf-files'
teardown() {
    rm testfile_send testfile_recv script.log || true
}

setup() {
    if [ -n "$AWS_ACCESS_KEY_ID" ]; then
        echo "AWS_ACCESS_KEY_ID must be set"
    fi

    if [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
        echo "AWS_SECRET_ACCESS_KEY must be set"
    fi

    rm testfile_send testfile_recv script.log || true
}

@test "gcp: basic use case: upload 10mb file then download it" {
    openssl rand $((1024*1024*10)) > testfile_send

    OUTPUT=$(./qf < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf "$ID" > testfile_recv

    DUPLICATE_HASH=$(md5sum testfile_send testfile_recv | awk '{ print $1 }' | uniq -c | awk '{ print $1 }')
    [ "$DUPLICATE_HASH" == 2 ]
}

@test "aws: basic use case: upload 10mb file then download it" {
    openssl rand $((1024*1024*10)) > testfile_send

    OUTPUT=$(./qf-aws < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf-aws "$ID" > testfile_recv

    DUPLICATE_HASH=$(md5sum testfile_send testfile_recv | awk '{ print $1 }' | uniq -c | awk '{ print $1 }')
    [ "$DUPLICATE_HASH" == 2 ]
}

@test "aws: basic use case: upload tiny file (1kb) then download it" {
    openssl rand $((1024)) > testfile_send

    OUTPUT=$(./qf-aws < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf-aws "$ID" > testfile_recv

    DUPLICATE_HASH=$(md5sum testfile_send testfile_recv | awk '{ print $1 }' | uniq -c | awk '{ print $1 }')
    [ "$DUPLICATE_HASH" == 2 ]
}

@test "gcp: basic use case: upload tiny file (1kb) then download it" {
    openssl rand $((1024)) > testfile_send

    OUTPUT=$(./qf < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf "$ID" > testfile_recv

    DUPLICATE_HASH=$(md5sum testfile_send testfile_recv | awk '{ print $1 }' | uniq -c | awk '{ print $1 }')
    [ "$DUPLICATE_HASH" == 2 ]
}


@test "aws: basic use case: start downloading file before it's completely uploaded" {
    openssl rand $((1024*1024*25)) > testfile_send
    nohup cat testfile_send | ./qf-aws > script.log || true

    sleep 3
    ID=$(cat script.log | awk '{ print $2 }')
    echo $ID
    ./qf-aws "$ID" > testfile_recv

    DUPLICATE_HASH=$(md5sum testfile_send testfile_recv | awk '{ print $1 }' | uniq -c | awk '{ print $1 }')
    [ "$DUPLICATE_HASH" == 2 ]
}


@test "aws: make sure file is not deleted when passing '-k' argument" {
    openssl rand $((1024)) > testfile_send

    OUTPUT=$(./qf-aws < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf-aws -k "$ID" > /dev/null

    OUTPUT=$(s3cmd ls s3://$BUCKET)
    [ "$OUTPUT" != "" ]
}

@test "aws: make sure files are deleted when passing '-d' argument" {
    openssl rand $((1024)) > testfile_send

    ./qf-aws < testfile_send
    ./qf-aws < testfile_send

    OUTPUT=$(s3cmd ls s3://$BUCKET)
    [ "$OUTPUT" != "" ]

    ./qf-aws -d

    OUTPUT=$(s3cmd ls s3://$BUCKET)
    [ "$OUTPUT" == "" ]
}



@test "aws: make sure file is deleted from aws after successful download" {
    s3cmd rm "s3://$BUCKET/*"
    openssl rand $((1024)) > testfile_send

    OUTPUT=$(./qf-aws < testfile_send)
    ID=$(echo "$OUTPUT" | awk '{ print $2 }')

    ./qf-aws "$ID" > /dev/null

    OUTPUT=$(s3cmd ls s3://$BUCKET)
    [ "$OUTPUT" == "" ]
}
