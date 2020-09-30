.PHONY: clean # flag dizendo ao Make que nenhum arquivo vai ser gerado quando chamar "make clean"
clean:
	rm -rf k8s-metadata
all:
	go build -o k8s-metadata main.go
