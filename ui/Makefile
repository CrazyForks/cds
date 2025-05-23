UNAME := $(shell uname)
SHA512 := $(if ifeq ${UNAME} "Darwin",shasum -a 512,sha512sum)
VERSION := $(if ${CDS_VERSION},${CDS_VERSION},snapshot)

clean:
	@rm -rf dist
	@rm -f ui.tar.gz
	@rm -rf node_modules
	@find semantic -name "*" -not -path "*/site/globals/site.variables" -delete
	@rm -rf semantic\\

NG = node_modules/@angular/cli/bin/ng
$(NG):
	npm ci

stats: $(NG)
	@node --max-old-space-size=3048 node_modules/@angular/cli/bin/ng build ui-ng2 --configuration production --stats-json

ngbuild: $(NG)
	@node --max-old-space-size=3048 node_modules/@angular/cli/bin/ng build ui-ng2 --configuration production
	@node --max-old-space-size=3048 node_modules/@angular/cli/bin/ng build workflow-graph --configuration production

INDEX = dist/browser/index.tmpl
$(INDEX):
	$(MAKE) ngbuild
	@cd dist/browser && mv index.html index.tmpl

FILES_UI = dist/browser/FILES_UI
$(FILES_UI): $(INDEX)
	$(info sha512 = ${SHA512})
	touch $(FILES_UI)
	cd dist/browser/ && for i in `ls -p | grep -v /|grep -v FILES_UI`; do echo "$$i;`${SHA512} $$i|cut -d ' ' -f1`" >> FILES_UI; done;

build: $(FILES_UI) $(INDEX) ui.tar.gz

ui.tar.gz:
	rm -rf dist/dist
	cp -R dist/browser dist/dist
	tar cfz ui.tar.gz -C dist dist

lintfix:
	./node_modules/.bin/ng lint --fix
