# Chamgo-qt

PLATFORMS=linux windows_64_static windows_32_static darwin_static_base

help:
	@echo "make build     # build all platforms ($(PLATFORMS)) "
	@$(foreach platform, $(PLATFORMS), printf "make %-24s # compile %s \\n" $(platform) $(platform);)

define compile_platform
.PHONY: $(1)

$(1): cleandeploy vendor
	@echo "Compiling for $(1) ..."
	docker pull therecipe/qt:$(1) > /dev/null
	qtdeploy -docker build $(1)

endef

$(foreach platform, $(PLATFORMS), $(eval $(call compile_platform,$(platform))))

build: linux windows_64_static windows_32_static darwin_static_base

vendor:
	@echo "Vendoring ..."
	@go mod vendor

cleandeploy:
	@echo "Cleaning ..."
	@rm -fr deploy

.PHONY: build vendor cleandeploy
