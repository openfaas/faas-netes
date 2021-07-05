# Testing the Chart locally

1.  Install Kafka in a separate namespace

    ```sh
    $ helm repo add incubator https://charts.helm.sh/incubator
    $ helm upgrade kf incubator/kafka \
      --install \
      --namespace openfaas \
      --set imageTag=4.1.3 \
      --set persistence.enabled=false

    $ kubectl -n openfaas apply -f - <<EOF
    apiVersion: v1
    kind: Pod
    metadata:
      name: client
      namespace: openfaas
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

    * Make sure that the Kafka broker replicas is odd and at least 3, otherwise messages will be
    processed because a quorum can not be reached. You may see this error as a symptom of a missing
    quorum https://github.com/bsm/sarama-cluster/issues/118

    * Version 5+ of the `confluentinc/cp-kafka` Docker image is not compatible with the connector at this time.

    * Verify the topics available:
        ```sh
        kubectl -n openfaas exec -it client -- kafka-topics --zookeeper kf-zookeeper:2181 --list
        ```

    * Create a partition and topic for faas-request
        ```sh
        kubectl -n openfaas exec -it client -- kafka-topics --zookeeper kf-zookeeper:2181 --topic faas-request --create --partitions 1 --replication-factor 1

    * Publish a message and play back the messages after:
        ```sh
        echo "test msg" | kubectl -n openfaas exec -it client -- kafka-console-producer --broker-list kf-kafka-headless:9092 --topic faas-request

        kubectl -n openfaas exec -it client -- kafka-console-consumer --bootstrap-server kf-kafka:9092 --topic faas-request --from-beginning
        ```

2) Install OpenFaaS

  ```sh
  arkade install openfaas
  ```

  Or

   ```sh
   helm upgrade openfaas --install openfaas/openfaas \
   --namespace openfaas  \
   --set basic_auth=true \
   --set functionNamespace=openfaas-fn
   ```

3) Install the connector using the arkade app:

   ```sh
   arkade install kafka-connector \
   --broker-host=kf-kafka:9092 \
   --license-file $HOME/.openfaas/LICENSE \
   --topics faas-request \
   --set image=alexellis/kafka-connector-pro:0.4.1-5-g551d268-dirty-amd64
  ```

4) Or install the connector from the faas-netes repo:

   ```sh
   cd charts
   helm upgrade kafka-connector . \
       --install \
       --namespace openfaas \
       --set broker_host=kf-kafka.kafka
   ```

4) Trigger a function from a topic:

   a. Deploy figlet

   ```sh
   faas store deploy figlet --annotation topic="faas-request" --gateway $GATEWAY
   ```

   b. Write a msg with the `client` to the topic `faas-request`:

   ```sh
   echo "This is a UNIX system" | kubectl -n openfaas exec -it client -- kafka-console-producer --broker-list kf-kafka-headless:9092 --topic faas-request
   ```

   c. Check the connector logs:

   ```sh
   kubectl -n openfaas logs deploy/kafka-connector -f
   ```

5. If you need to reset your environment:

   ```sh
   faas remove figlet --gateway $GATEWAY
   kubectl -n openfaas delete pod/client

   helm delete --purge kafka-connector
   helm delete --purge kf
   ```

   Optionally, also remove OpenFaaS:
   ```sh
   helm delete --purge openfaas
   ```
