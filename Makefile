
.PSEUDO: install push

install:
	go build --tags winfsp,osusergo,netgo -o iphonefs -ldflags "-w -s" .
	mv iphonefs /usr/local/bin

push:
	git push "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push "https://github.com/systemmonkey42/iphonefs" "develop:develop"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:develop"

package:
	env GOOS="linux" GOARCH="amd64" go build --tags winfsp,osusergo,netgo -o iphonefs -ldflags "-w -s" .
	env GOOS="windows" GOARCH="amd64" go build --tags winfsp,osusergo,netgo -o iphonefs.exe -ldflags "-w -s" .
