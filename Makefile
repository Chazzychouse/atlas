LAST_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
MAJOR := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f1)
MINOR := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f2)
PATCH := $(shell echo $(LAST_TAG) | sed 's/v//' | cut -d. -f3)

.PHONY: release release-minor release-major

## release: bump patch version, tag, and push (v0.0.1 -> v0.0.2)
release:
	@NEXT=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))); \
	echo "$(LAST_TAG) -> $$NEXT"; \
	git tag $$NEXT && git push origin $$NEXT

## release-minor: bump minor version, tag, and push (v0.1.0 -> v0.2.0)
release-minor:
	@NEXT=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0; \
	echo "$(LAST_TAG) -> $$NEXT"; \
	git tag $$NEXT && git push origin $$NEXT

## release-major: bump major version, tag, and push (v1.0.0 -> v2.0.0)
release-major:
	@NEXT=v$(shell echo $$(($(MAJOR)+1))).0.0; \
	echo "$(LAST_TAG) -> $$NEXT"; \
	git tag $$NEXT && git push origin $$NEXT
