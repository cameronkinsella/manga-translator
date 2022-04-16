#!/bin/bash

gcloud projects list
echo Enter the PROJECT_ID that you want to use:
read PROJECT_ID
gcloud config set project $PROJECT_ID

gcloud iam service-accounts create mtl-service-account \
    --display-name="cloud-vision"

gcloud iam service-accounts keys create "mtl-service-account-keys.json" \
    --iam-account=mtl-service-account@$PROJECT_ID.iam.gserviceaccount.com

cloudshell download ./mtl-service-account-keys.json
