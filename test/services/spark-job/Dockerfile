ARG SPARK_IMAGE=spark:3.5.0-scala2.12-java11-ubuntu

FROM $SPARK_IMAGE

COPY entrypoint.sh .

ENTRYPOINT ["/opt/spark/work-dir/entrypoint.sh"]
