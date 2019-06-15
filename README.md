# [WIP]thanos-operator

## Create gcs iam service-account

```sh
PROJECT_ID=$(gcloud config get-value core/project)
SERVICE_ACCOUNT_NAME="thanos-demo-gcs"

gcloud iam service-accounts create ${SERVICE_ACCOUNT_NAME} \
  --quiet \
  --display-name "thanos demo gcs"

gcloud projects add-iam-policy-binding ${SERVICE_ACCOUNT_NAME} \
  --quiet \
  --member="serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/storage.objectCreator

gcloud projects add-iam-policy-binding ${SERVICE_ACCOUNT_NAME} \
  --quiet \
  --member "serviceAccount:${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role roles/storage.objectViewer

gcloud iam service-accounts keys create ${SERVICE_ACCOUNT_NAME}.json \
  --quiet \
  --iam-account ${SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com
```

## Create secret from iam json file

```sh
kubectl create secret generic ${SERVICE_ACCOUNT_NAME} \
  --from-file=${SERVICE_ACCOUNT_NAME}.json=${SERVICE_ACCOUNT_NAME}.json
```

## Create bucket

```sh
BUCKET_NAME="orangesys-thanos-demo"
gsutil mb gs://${BUCKET_NAME}/
```

