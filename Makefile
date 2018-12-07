VERSION=1.0.0.0
IMAGE=us.gcr.io/kafka-pubsub-emulator/kafka-pubsub-emulator
JAR=./kafka-pubsub-emulator/target/kafka-pubsub-emulator-$(VERSION).jar
SOURCES_DIR=kafka-pubsub-emulator/src/main/java/com/google/cloud/partners/pubsub/kafka/
SOURCES=$(shell find $(SOURCES_DIR) -type f -name '*.java')

KAFKA_RELEASE=kafka1
EMULATOR_RELEASE=emu1

kafka-pubsub-emulator:
	git clone https://github.com/riccardomc/kafka-pubsub-emulator

%.jar: $(SOURCES) kafka-pubsub-emulator
	@echo $(SOURCES)
	(cd kafka-pubsub-emulator ; mvn package -DskipTests=True)

.PHONY: build
build: $(JAR)
	eval $$(minikube docker-env) ;\
	docker build --build-arg version=$(VERSION) -t $(IMAGE):$(VERSION) ./kafka-pubsub-emulator

.PHONY: helm-deploy
deploy:
	helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
	helm repo add containersolutions https://containersolutions.github.io/helm-charts/
	helm repo update
	helm upgrade -i $(KAFKA_RELEASE) incubator/kafka
	helm upgrade -i --set emulatorConfig.kafka.bootstrapServers='$(KAFKA_RELEASE)-0.$(KAFKA_RELEASE)-headless:9092\,$(KAFKA_RELEASE)-1.$(KAFKA_RELEASE)-headless:9092\,$(KAFKA_RELEASE)-2.$(KAFKA_RELEASE)-headless:9092' \
		$(EMULATOR_RELEASE) \
		containersolutions/kafka-pubsub-emulator

.PHONY: tests
tests:
	go test -v ./tests/... -count=1

.PHONY: logs
logs:
	@while (true) ; do kubectl get pods | grep $(EMULATOR_RELEASE) | awk '{print $$1}' | xargs kubectl logs -f ; slee 0.5 ; done

.PHONY: refresh
refresh: build
	kubectl get pods | grep $(EMULATOR_RELEASE) | awk '{print $$1}' | xargs kubectl delete pod 

