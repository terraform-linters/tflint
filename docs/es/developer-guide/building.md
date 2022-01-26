# Construir TFLint

Es necesario tener la versión 1.17 o superior de GO para poder construir TFLint desde el código fuente. Clone el código fuente y ejecute el comando `make`. El binario construido se colocará en el directorio `dist`.

```console
$ git clone https://github.com/terraform-linters/tflint.git
$ cd tflint
$ make
mkdir -p dist
go build -v -o dist/tflint
```

## Ejecutar los tests

Si cambia el código, asegúrese de que las pruebas que añada y las ya existentes pasen correctamente:

```console
$ make test
```

## Ejecutar los tests E2E

Puede comprobar el comportamiento real del CLI ejecutando las pruebas E2E. Dado que las pruebas E2E utilizan el comando `tflint`, es necesario añadir esta ruta a la variable de entorno `$PATH` de esta forma el binario construido mediante el comando `go install` pueda ser referenciado.

```console
$ make e2e
```
