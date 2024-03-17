#!/bin/bash

K8S_API_TOKEN="$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"
K8S_API_CERT="/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

SPARK_IMAGE=spark:3.5.0-scala2.12-java11-ubuntu

echo "Starting job = $JOB_NAME"

/opt/spark/bin/./spark-submit \
    --master k8s://https://kubernetes.default.svc:443 \
    --deploy-mode client \
    --name $JOB_NAME \
    --class org.apache.spark.examples.SparkPi \
    --conf spark.executor.instances=1 \
    --conf spark.driver.extraJavaOptions="--add-exports java.base/sun.nio.ch=ALL-UNNAMED" \
    --supervise \
    --conf spark.driver.host=$SPARK_HOSTNAME \
    --conf spark.kubernetes.executor.podNamePrefix=$JOB_NAME \
    --conf spark.kubernetes.container.image=$SPARK_IMAGE \
    --conf spark.kubernetes.authenticate.driver.serviceAccountName=$SERVICE_ACCOUNT_NAME \
    --conf spark.kubernetes.authenticate.executor.serviceAccountName=$SERVICE_ACCOUNT_NAME \
    --conf spark.kubernetes.authenticate.submission.oauthToken=$K8S_API_TOKEN \
    --conf spark.kubernetes.authenticate.submission.caCertFile=$K8S_API_CERT \
    --conf spark.ui.reverseProxy=true \
    file:///opt/spark/examples/jars/spark-examples_2.12-3.5.0.jar "$1"
