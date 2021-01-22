NAME=chamgo-qt

$(NAME): dimage vendor 
	qtdeploy -docker build desktop

dimage:
	docker pull therecipe/qt:linux

vendor:
	go mod vendor

.PHONY: vendor dimage
