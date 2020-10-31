

IS_DOCKER_INSTALLED = $(shell which docker >> /dev/null 2>&1; echo $$?)
# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')
ORG?=litmuschaos
ADMISSION_CONTROLLERS_REPO_NAME?=admission-controllers
# Specify the name for the binaries
WEBHOOK=admission-controllers
# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

.PHONY: all
all: deps gotasks admission-controllers-image

.PHONY: deps
deps: _build_check_docker godeps

.PHONY: godeps
godeps:
	@echo ""
	@echo "INFO:\tverifying dependencies for admission controller build ..."
	@go get  -v golang.org/x/lint/golint
	@go get  -v golang.org/x/tools/cmd/goimports

.PHONY: _build_check_docker
_build_check_docker:
	@if [ $(IS_DOCKER_INSTALLED) -eq 1 ]; \
		then echo "" \
		&& echo "ERROR:\tdocker is not installed. Please install it before build." \
		&& echo "" \
		&& exit 1; \
		fi;

.PHONY: test
test:
	@echo "------------------"
	@echo "--> Run Go Test"
	@echo "------------------"
	@go test ./... -coverprofile=coverage.txt -v

.PHONY: gotasks
gotasks: format lint

.PHONY: format
format:
	@echo "------------------"
	@echo "--> Running go fmt"
	@echo "------------------"
	@go fmt $(PACKAGES)

.PHONY: lint
lint:
	@echo "------------------"
	@echo "--> Running golint"
	@echo "------------------"
	@golint $(PACKAGES)
	@echo "------------------"
	@echo "--> Running go vet"
	@echo "------------------"
	@go vet $(PACKAGES)



.PHONY: admission-controllers-image
admission-controllers-image:
	@echo "----------------------------"
	@echo -n "--> admission-controllers image: "
	@echo "${ORG}/${ADMISSION_CONTROLLERS_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-controllers/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${ORG}/${ADMISSION_CONTROLLERS_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}