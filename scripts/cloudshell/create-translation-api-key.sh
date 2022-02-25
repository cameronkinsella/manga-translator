#!/bin/bash

gcloud projects list
echo Enter the PROJECT_ID that you want to use:
read PROJECT_ID
gcloud config set project $PROJECT_ID

gcloud alpha services api-keys create --display-name="MTL Cloud Translation" --api-target=service=translate.googleapis.com

export KEY_NAME="$(gcloud alpha services api-keys list --filter="displayName='MTL Cloud Translation'" --format="value(name)")"

echo Your Cloud Translation API key:
gcloud alpha services api-keys get-key-string $KEY_NAME
