# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')
HUB_USER?=rahulchheda1997
ADMISSION_SERVER_REPO_NAME?=admission-server
# Specify the name for the binaries
WEBHOOK=admission-server
# Specify the date o build
BUILD_DATE = $(shell date +'%Y%m%d%H%M%S')

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

.PHONY: all
all: admission-server-image

.PHONY: admission-server-image
admission-server-image:
	@echo "----------------------------"
	@echo -n "--> admission-server image "
	@echo "${HUB_USER}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-server/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${HUB_USER}/${ADMISSION_SERVER_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}