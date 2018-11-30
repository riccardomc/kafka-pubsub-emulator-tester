# Test Kafka Pub/Sub Emulator

This repository contains a set of tests for the functionality of 
[GCP Kafka Pub/Sub Emulator](https://github.com/GoogleCloudPlatform/kafka-pubsub-emulator)

Some instructions on how to run the tests on
[minikube](https://kubernetes.io/docs/setup/minikube/) are provided in the
remainder of this README file.

This thing is not supposed to be user friendly. I just need to run some tests.

## Deploy

The deployment step uses [helm](https://helm.sh/). The emulator will need a
Kubernetes cluster and a Kafka cluster.  Here's an example using minikube:

```
minikube start
helm init
make deploy
```

This will start minikube, deploy a Kafka cluster named `kafka1` and an pubsub
emulator `emu1` attached to it and exposed using a `NodePort` service. Here it
is:

```
% minikube service list                         
|-------------|----------------------------|-----------------------------|
|  NAMESPACE  |            NAME            |             URL             |
|-------------|----------------------------|-----------------------------|
| default     | emu1-kafka-pubsub-emulator | http://192.168.99.100:32722 |
| default     | kafka1                     | No node port                |
| default     | kafka1-headless            | No node port                |
| default     | kafka1-zookeeper           | No node port                |
| default     | kafka1-zookeeper-headless  | No node port                |
| default     | kubernetes                 | No node port                |
| kube-system | kube-dns                   | No node port                |
| kube-system | kubernetes-dashboard       | No node port                |
| kube-system | tiller-deploy              | No node port                |
|-------------|----------------------------|-----------------------------|
```

## Run the tests

Tests are written in [go](https://golang.org/) and [go
modules](https://github.com/golang/go/wiki/Modules) so you'll need go 1.11 at
least.

The emulator service can be referenced from outside the minikube cluster by
setting the `PUBSUB_EMULATOR_HOST` environment variable. A set of tests can be
run against the emulator by:

```
export PUBSUB_EMULATOR_HOST=192.168.99.100:32722
make tests
```

## Deploy Pub/Sub emulator with local changes

You can build emulator locally using the provided Makefile:
```
make refresh
```
This will build the emulator application, create a docker image in minikube
context and kill the pod currently deployed so it will be restarted using the
latest image.

This shows you the logs of the emulator and restarts when the pod is refreshed:
```
make logs
```

## Manual testing of the kafka cluster
You don't really need to do this, but if you want to verify what's happening on
the cluster, here's how:

```
cat < EOF > ./kafka-testclient.yaml
apiVersion: v1
kind: Pod
metadata:
  name: testclient
  namespace: kafka
spec:
  containers:
  - name: kafka
    image: solsson/kafka:0.11.0.0
    command:
      - sh
      - -c
      - "exec tail -f /dev/null"
EOF
kubectl apply -f ./kafka-testclient.yaml
kubectl exec -ti testclient -- /bin/bash
```

