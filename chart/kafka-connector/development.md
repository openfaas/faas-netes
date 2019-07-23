# Testing the Chart locally

1.  Install Kafka in a separate namespace

    ```sh
    $ helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
    $ kubectl create ns kafka
    $ helm upgrade kf incubator/kafka \
        --install \
        --namespace kafka \
        --set imageTag=4.1.3 \
        --set persistence.enabled=false
    $ kubectl -n kafka apply -f - <<EOF
    apiVersion: v1
    kind: Pod
    metadata:
      name: testclient
      namespace: kafka
    spec:
      containers:
      - name: kafka
        image: confluentinc/cp-kafka:5.0.1
        command:
        - sh
        - -c
        - "exec tail -f /dev/null"
    EOF
    ```

    * Make sure that the kafka broker replicas is odd and at least 3, otherwise messages will be
    processed because a quorum can not be reached. You may see this error as a symptom of a missing
    quorum https://github.com/bsm/sarama-cluster/issues/118

    * Verion 5+ of the `confluentinc/cp-kafka` Docker image is not compatible with the Connector

    * Testing kafka with the test client

        ```sh
        kubectl -n kafka exec -it testclient -- kafka-topics --zookeeper kf-zookeeper:2181 --list
        kubectl -n kafka exec -it testclient -- kafka-topics --zookeeper kf-zookeeper:2181 --topic faas-request --create --partitions 1 --replication-factor 1

        echo "test msg" | kubectl -n kafka exec -it testclient -- kafka-console-producer --broker-list kf-kafka-headless:9092 --topic faas-request

        kubectl -n kafka exec -it testclient -- kafka-console-consumer --bootstrap-server kf-kafka:9092 --topic faas-request --from-beginning
        ```

2) Install OpenFaaS

   ```sh
   helm upgrade openfaas --install openfaas/openfaas \
   --namespace openfaas  \
   --set basic_auth=true \
   --set functionNamespace=openfaas-fn
   ```

3) Install the Connector from the local directly

   ```sh
   helm upgrade kafka-connector . \
       --install \
       --namespace openfaas \
       --set broker_host=kf-kafka.kafka
   ```

4) Testing

   a. deploy figlet

   ```sh
   faas store deploy figlet --annotation topic="faas-request" --gateway $GATEWAY
   ```

   b. write a msg with the `testclient`:

   ```sh
   echo "test message" | kubectl -n kafka exec -it testclient -- kafka-console-producer --broker-list kf-kafka-headless:9092 --topic faas-request
   ```

   c. check the connector logs:

   ```sh
   kubectl -n openfaas logs $(kubectl -n openfaas get po -l "app.kubernetes.io/component=kafka-connector" -o name | cut -d'/' -f2) -f
   ```

5. Delete everything

   ```sh
   faas remove figlet --gateway $GATEWAY
   kubectl -n kafka delete po testclient
   helm delete --purge kafka-connector
   helm delete --purge kf
   helm delete --purge openfaas
   ```
