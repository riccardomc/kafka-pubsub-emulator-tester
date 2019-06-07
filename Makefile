VERSION ?= 0.1.0
IMAGE=us.gcr.io/kafka-pubsub-emulator/kafka-pubsub-emulator
JAR=./kafka-pubsub-emulator/target/kafka-pubsub-emulator-$(VERSION).jar
SOURCES_DIR=kafka-pubsub-emulator/src/main/java/com/google/cloud/partners/pubsub/kafka/
SOURCES=$(shell find $(SOURCES_DIR) -type f -name '*.java')

NAMESPACE=default

KAFKA_RELEASE=kafka1
KAFKA_CHART=incubator/kafka
KAFKA_CHART_VERSION=0.13.5

EMULATOR_RELEASE=emu1
EMULATOR_NAMESPACE=emu
EMULATOR_CHART=/home/rmc/Development/Projects/AtoS/helm-charts/charts/kafka-pubsub-emulator
EMULATOR_CHART_VERSION=0.2.0

SAMPLE_CONFIG_DIR=$(shell pwd)/kafka-pubsub-emulator/demo/benchmark/config

FORK=https://github.com/riccardomc/kafka-pubsub-emulator

kafka-pubsub-emulator:
	git clone $(FORK)

%.jar: $(SOURCES) kafka-pubsub-emulator
	@echo $(SOURCES)
	(cd kafka-pubsub-emulator ; mvn package -DskipTests=True)

.PHONY: build
build: $(JAR)
	eval $$(minikube docker-env) ;\
	docker build --build-arg version=$(VERSION) -t $(IMAGE):$(VERSION) ./kafka-pubsub-emulator

.PHONY: run
run:
	docker run --mount type=bind,src=$(SAMPLE_CONFIG_DIR),dst=/etc/config \
		$(IMAGE):$(VERSION) \
		-c /etc/config/config.json \
		-p /etc/config/pubsub.json


.PHONY: deploy
deploy:
	helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
	helm repo add containersolutions https://containersolutions.github.io/helm-charts/
	helm repo update
	helm upgrade -i \
		--namespace $(NAMESPACE) \
		--set external.enabled="true" \
		--set external.loadBalancerIP[0]="192.168.99.103" \
		--set external.loadBalancerIP[1]="192.168.99.103" \
		--set external.loadBalancerIP[2]="192.168.99.103" \
		$(KAFKA_RELEASE) \
		$(KAFKA_CHART) \
		--version $(KAFKA_CHART_VERSION)
	helm upgrade -i \
		--namespace $(NAMESPACE) \
		--set server.kafka.bootstrapServers[0]="$(KAFKA_RELEASE)-0.$(KAFKA_RELEASE)-headless.${NAMESPACE}:9092" \
		--set server.kafka.bootstrapServers[1]="$(KAFKA_RELEASE)-1.$(KAFKA_RELEASE)-headless.${NAMESPACE}:9092" \
		--set server.kafka.bootstrapServers[2]="$(KAFKA_RELEASE)-2.$(KAFKA_RELEASE)-headless.${NAMESPACE}:9092" \
		$(EMULATOR_RELEASE) \
		$(EMULATOR_CHART) \
		--version $(EMULATOR_CHART_VERSION)

.PHONY: tests
tests:
	go test -v ./tests/... -count=1

.PHONY: logs
logs:
	@while (true) ; do kubectl -n $(EMULATOR_NAMESPACE) get pods | grep $(EMULATOR_RELEASE)-kafka-pubsub-emulator | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) logs -f ; slee 0.5 ; done

.PHONY: refresh
refresh: build
	kubectl -n $(EMULATOR_NAMESPACE) get pods | grep $(EMULATOR_RELEASE)-kafka-pubsub-emulator | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) delete pod

.PHONY: deepclean
deepclean:
	helm delete --purge $(KAFKA_RELEASE) $(EMULATOR_RELEASE) || true
	kubectl -n $(EMULATOR_NAMESPACE) get persistentvolumeclaims | grep $(EMULATOR_RELEASE) | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) delete persistentvolumeclaim || true
	kubectl -n $(EMULATOR_NAMESPACE) get persistentvolumeclaims | grep $(KAFKA_RELEASE) | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) delete persistentvolumeclaim || true
	kubectl -n $(EMULATOR_NAMESPACE) get persistentvolume | grep $(EMULATOR_RELEASE) | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) delete persistentvolume || true
	kubectl -n $(EMULATOR_NAMESPACE) get persistentvolume | grep $(KAFKA_RELEASE) | awk '{print $$1}' | xargs kubectl -n $(EMULATOR_NAMESPACE) delete persistentvolume || true
