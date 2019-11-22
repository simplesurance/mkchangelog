GIT_COMMIT := $(shell git rev-parse HEAD)
VERSION := $(shell cat ver)

LDFLAGS := "-X main.buildCommit=$(GIT_COMMIT) \
	    -X main.buildVersion=$(VERSION)"

TARFLAGS := --sort=name --mtime='2018-01-01 00:00:00' --owner=0 --group=0 --numeric-owner
DOCKER_REPO := simplesurance/mkchangelog
DOCKER_ARG_TAGS := -t $(DOCKER_REPO):latest -t $(DOCKER_REPO):$(VERSION)

.PHONY: all
all:
	go build -ldflags=$(LDFLAGS) -o mkchangelog cmd/mkchangelog/main.go

.PHONY: clean
clean:
	@rm -rf mkchangelog dist/

.PHONY: dist/darwin_amd64/mkchangelog
dist/darwin_amd64/mkchangelog:
	$(info * building $@)
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build \
		-ldflags=$(LDFLAGS) -o "$@" cmd/mkchangelog/main.go

	$(info * creating $(@D)/mkchangelog-darwin_amd64-$(VERSION).tar.xz)
	@tar $(TARFLAGS) -C $(@D) -cJf $(@D)/mkchangelog-darwin_amd64-$(VERSION).tar.xz $(@F)

	$(info * creating $(@D)/mkchangelog-darwin_amd64-$(VERSION).tar.xz.sha256)
	@(cd $(@D) && sha256sum mkchangelog-darwin_amd64-$(VERSION).tar.xz > mkchangelog-darwin_amd64-$(VERSION).tar.xz.sha256)

.PHONY: dist/linux_amd64/mkchangelog
dist/linux_amd64/mkchangelog:
	$(info * building $@)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-ldflags=$(LDFLAGS) -o "$@" cmd/mkchangelog/main.go

	$(info * creating $(@D)/mkchangelog-linux_amd64-$(VERSION).tar.xz)
	@tar $(TARFLAGS) -C $(@D) -cJf $(@D)/mkchangelog-linux_amd64-$(VERSION).tar.xz $(@F)

	$(info * creating $(@D)/mkchangelog-linux_amd64-$(VERSION).tar.xz.sha256)
	@(cd $(@D) && sha256sum mkchangelog-linux_amd64-$(VERSION).tar.xz > mkchangelog-linux_amd64-$(VERSION).tar.xz.sha256)

.PHONY: docker_image
docker_image: dist/linux_amd64/mkchangelog
	@mkdir -p docker/files
	@cp dist/linux_amd64/mkchangelog docker/files/
	( cd docker && docker build  $(DOCKER_ARG_TAGS) . )

.PHONY: dirty_worktree_check
dirty_worktree_check:
	@if ! git diff-files --quiet || git ls-files --other --directory --exclude-standard | grep ".*" > /dev/null ; then \
		echo "remove untracked files and changed files in repository before creating a release, see 'git status'"; \
		exit 1; \
		fi

.PHONY: release
release: clean dirty_worktree_check dist/linux_amd64/mkchangelog dist/darwin_amd64/mkchangelog docker_image
	@echo
	@echo next steps:
	@echo - git tag v$(VERSION)
	@echo - git push --tags
	@echo - upload $(ls dist/*/*.tar.xz) files
	@echo - "push docker image - docker push $(DOCKER_REPO):latest && docker push $(DOCKER_REPO):$(VERSION)"
