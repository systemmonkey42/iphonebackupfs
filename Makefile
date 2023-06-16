
.PSEUDO: install push

build=$(shell git describe --tags)

install:
	go build --tags winfsp,osusergo,netgo -o iphonebackupfs -ldflags "-w -s" .
	mv iphonebackupfs /usr/local/bin

push:
	git push "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push "https://github.com/systemmonkey42/iphonefs" "develop:develop"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:develop"

package:
	env GOOS="linux" GOARCH="amd64" go build --tags winfsp,osusergo,netgo -o iphonebackupfs -ldflags "-w -s" .
	env GOOS="windows" GOARCH="amd64" go build --tags winfsp,osusergo,netgo -o iphonebackupfs.exe -ldflags "-w -s" .
	zip -9 iphonebackupfs-windows-$(build)-x86_64.zip iphonebackupfs.exe README.md
	tar cvfz iphonebackupfs-linux-$(build)-amd64.tgz iphonebackupfs README.md
