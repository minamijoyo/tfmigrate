#!/bin/bash
awslocal s3api wait bucket-exists --bucket "$S3_BUCKET"
