
.PSEUDO: install push

build=$(shell git describe --tags)

install: linux
	mv iphonebackupfs /usr/local/bin

linux:
	env GOOS="linux" GOARCH="amd64" go build -tags winfsp,osusergo,netgo -o iphonebackupfs -ldflags "-w -s" .

windows:
	env CC='/usr/bin/x86_64-w64-mingw32-gcc-win32' CGO_CFLAGS="-O2 -g -I${PWD}/../winfsp/inc/fuse" CGO_ENABLED=1 GOOS="windows" GOARCH="amd64" go build --tags winfsp,osusergo,netgo -o iphonebackupfs.exe -ldflags "-w -s" .

push:
	git push "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push "https://github.com/systemmonkey42/iphonefs" "develop:develop"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:develop"

windows_zip: windows
	zip -9 iphonebackupfs-windows-$(build)-x86_64.zip iphonebackupfs.exe README.md

linux_tar: linux
	tar cvfz iphonebackupfs-linux-$(build)-amd64.tgz iphonebackupfs README.md

package: linux_tar windows_zip
