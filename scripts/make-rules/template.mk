

.PHONY: template.build
template.build: $(addprefix template.build., $(addprefix $(subst /,_,$(IMAGE_PLAT))., $(IMAGES)))

.PHONY: template.build.%
template.build.%:
	$(eval COMMAND := $(word 2,$(subst ., ,$*)))
	$(eval PLATFORM := $(word 1,$(subst ., ,$*)))
	$(eval OS := $(word 1,$(subst _, ,$(PLATFORM))))
	$(eval ARCH := $(word 2,$(subst _, ,$(PLATFORM))))
	$(eval IMAGE := $(COMMAND))
	$(eval IMAGE_TAG := $(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION))
	@echo "===========> Building kubernetes yaml for $(IMAGE), image:$(IMAGE_TAG)"
	@mkdir -p $(TMP_DIR)/$(IMAGE)
#	@cat $(ROOT_DIR)/build/docker/$(IMAGE)/Dockerfile\
#		| sed -e "s#BASE_IMAGE#$(BASE_IMAGE)#g" -e "s#IMAGE_COMMAND#$(IMAGE)$(GO_OUT_EXT)#g" \
#		 -e "s#IMAGE_CONFIG#$(IMAGE).yaml#g" >$(TMP_DIR)/$(IMAGE)/Dockerfile
#	@cp $(OUTPUT_DIR)/configs/$(IMAGE).yaml $(TMP_DIR)/$(IMAGE)/ || true
#	@cp $(OUTPUT_DIR)/platforms/$(IMAGE_PLAT)/$(IMAGE)$(GO_OUT_EXT) $(TMP_DIR)/$(IMAGE)/
#	@DST_DIR=$(TMP_DIR)/$(IMAGE) $(ROOT_DIR)/build/docker/$(IMAGE)/build.sh 2>/dev/null || true
#	$(eval BUILD_SUFFIX := $(_DOCKER_BUILD_EXTRA_ARGS) --pull -t $(REGISTRY_PREFIX)/$(IMAGE)-$(ARCH):$(VERSION) $(TMP_DIR)/$(IMAGE))
#	@if [ $(shell $(GO) env GOARCH) != $(ARCH) ] ; then \
#		$(MAKE) image.daemon.verify ;\
#		$(DOCKER) build --platform $(IMAGE_PLAT) $(BUILD_SUFFIX) ; \
#	else \
#		$(DOCKER) build $(BUILD_SUFFIX) ; \
#	fi
	@rm -rf $(TMP_DIR)/$(IMAGE)
