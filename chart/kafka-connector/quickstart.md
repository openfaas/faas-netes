# Kafka Pro connector quickstart

1.  Install Kafka in a separate namespace

    ```sh
    $ arkade install kafka

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
        kubectl -n openfaas exec -it client -- kafka-topics --zookeeper cp-helm-charts-cp-zookeeper.default:2181 --list
        ```

    * Create a partition and topic for faas-request
        ```sh
        kubectl -n openfaas exec -it client -- kafka-topics --zookeeper cp-helm-charts-cp-zookeeper.default:2181 --topic faas-request --create --partitions 1 --replication-factor 1
        ```

    * Publish a message and play back the messages after:
        ```sh
       echo "test msg" | kubectl -n openfaas exec -it client -- kafka-console-producer --broker-list cp-helm-charts-cp-kafka-headless.default:9092 --topic faas-request
       kubectl -n openfaas exec -it client -- kafka-console-consumer --bootstrap-server cp-helm-charts-cp-kafka.default:9092 --topic faas-request --from-beginning
        ```

2) Install OpenFaaS

  ```sh
  arkade install openfaas
  ```

  Or use helm

3) Create any secrets required

  Secrets are required for SASL or client certificate authentication, see the comments in the [values.yaml](values.yaml) file

4) Install the connector:

  Create `overrides.yaml` and configure as per comments in [values.yaml](values.yaml)

  Example for Kafka helm chart:

  ```yaml
  brokerHost: cp-helm-charts-cp-kafka.default:9092
  tls: false
  saslAuth: false

  caSecret: ""
  certSecret: ""
  keySecret: ""
  ```

  Example for Aiven cloud with client certificates:

  ```yaml
  brokerHost: kafka-202504b5-openfaas-910b.aivencloud.com:10905
  tls: true
  saslAuth: false
  
  caSecret: kafka-broker-ca
  certSecret: kafka-broker-cert
  keySecret: kafka-broker-key
  ```

  Example for Confluent Cloud with Let's Encrypt and SASL auth:

  ```yaml
  brokerHost: pkc-4r297.europe-west1.gcp.confluent.cloud:9092
  tls: true
  saslAuth: true
  
  caSecret: ""
  certSecret: ""
  keySecret: ""
  ```

  Use the helm chart:

   ```sh
   cd faas-netes/charts
   
   helm upgrade kafka-connector ./kafka-connector \
       --install \
       --namespace openfaas \
       --values ./overrides.yaml
   ```

  Or install with arkade:

   ```sh
   arkade install kafka-connector \
   --broker-host=cp-helm-charts-cp-kafka.default:9092 \
   --license-file $HOME/.openfaas/LICENSE \
   --topics faas-request
   ```

5) Trigger a function from a topic:

   a. Deploy figlet

   ```sh
   faas store deploy figlet --annotation topic="faas-request" --gateway $GATEWAY
   ```

   b. Write a msg with the `client` to the topic `faas-request`:

   ```sh
   echo "This is a UNIX system" | kubectl -n openfaas exec -it client -- kafka-console-producer --broker-list cp-helm-charts-cp-kafka-headless.default:9092 --topic faas-request
   ```

   c. Check the connector logs:

   ```sh
   kubectl -n openfaas logs deploy/kafka-connector -f
   ```

6. If you need to reset your environment:

   ```sh
   faas remove figlet --gateway $GATEWAY
   kubectl -n openfaas delete pod/client

   helm uninstall -n openfaas kafka-connector
   helm uninstall -n openfaas kf
   ```

   Optionally, also remove OpenFaaS:
   ```sh
   helm uninstall -n openfaas openfaas
   ```
