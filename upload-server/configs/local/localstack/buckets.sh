#!/usr/bin/env bash
awslocal s3 mb s3://test-bucket
awslocal sns create-topic --name report-topic
awslocal sqs create-queue --queue-name report-queue
awslocal sns subscribe --topic-arn "arn:aws:sns:us-east-1:000000000000:report-topic" --protocol sqs --notification-endpoint "arn:aws:sqs:us-east-1:000000000000:report-queue" --attributes RawMessageDelivery=true
