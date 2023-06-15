
.PSEUDO: install push

install:
	go build --tags winfsp,osusergo,netgo -o iphonefs -ldflags "-w -s" .
	mv iphonefs /usr/local/bin

push:
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:main"
	git push --tags "https://github.com/systemmonkey42/iphonefs" "develop:develop"
