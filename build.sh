#!/bin/bash
SCRIPT_NAME=$(basename $0)

function verify_go {
   if ! hash go 2>/dev/null; then
       echo "go binary not found!" 
       exit 1
   fi
}

function build_aws {
    echo "Building ./qf"

    go build -ldflags \
        "-X main.awsRegion=$1 -X main.bucket=$2 -X main.awsKey=$3 -X main.awsSecret=$4" \
        -o qf

    echo "Build successful!"
}

function build_gc {
    echo "Building ./qf"

    go build -ldflags \
        "-X main.googleCreds=$(base64 -w0 "$2") -X main.bucket=$1" \
        -o qf

    echo "Build successful!"
}

verify_go

if [ "$1" = "-h" ]; then
    printf "pass either 'aws' or 'gcp'"
    exit 0
fi


if [ "$1" = "aws" ]; then
    if [ "$2" == "" ] || [ "$3" == "" ] || [ "$4" == "" ] || [ "$5" == "" ]; then
        printf "pass AWS_REGION, AWS_BUCKET, KEY_ID and SECRET as arguments\n\nExample:\n\t./$SCRIPT_NAME aws 'us-east-1' 'qf-files' 'AKIAJADX1MDIVREDIN3A' '3VPyc0c2U12vFa8vBfVbqktaXUUpdzxWExVTZ'"
    else
        echo "building AWS client!"
        build_aws "$2" "$3" "$4" "$5"
    fi  
elif [ "$1" = "gcp" ]; then
    if [ "$2" != "" ] && [ -f "$3" ]; then
        echo "building Google Cloud client!"
        build_gc "$2" "$3"
    else
        printf "last argument should be Google Object Storage bucket followed by location of Google Cloud JSON credentials file"
        printf "\n\nExample:\n\t./$SCRIPT_NAME aws 'bucket-files' creds.json"
    fi
else
    echo "Pass 'aws' or 'gcp' as the first positional argument:"
    printf "\nExample: \n\t'./$SCRIPT_NAME aws' or './$SCRIPT_NAME gcp'"
    exit 1
fi
