# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')
HUB_USER?=litmuschaos
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
all: admission-controllers-image

.PHONY: admission-controllers-image
admission-controllers-image:
	@echo "----------------------------"
	@echo -n "--> admission-controllers image "
	@echo "${HUB_USER}/${ADMISSION_CONTROLLERS_REPO_NAME}:${IMAGE_TAG}"
	@echo "----------------------------"
	@PNAME=${WEBHOOK} CTLNAME=${WEBHOOK} sh -c "'$(PWD)/buildscripts/build.sh'"
	@cp bin/${WEBHOOK}/${WEBHOOK} buildscripts/admission-controllers/
	@cd buildscripts/${WEBHOOK} && sudo docker build -t ${HUB_USER}/${ADMISSION_CONTROLLERS_REPO_NAME}:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} .
	@rm buildscripts/${WEBHOOK}/${WEBHOOK}