# Escribir plugins

Si quiere añadir reglas personalizadas, puedes escribir plugins con nuevos conjuntos de reglas.

## Visión general

Los plugins son archivos binarios independientes y utilizan [go-plugin](https://github.com/hashicorp/go-plugin) para comunicase vía RPC con TFLint. TFLint ejecuta el archivo binario cuando se habilita el plugin, el proceso del plugin debe actuar como un servidor RPC para TFLint.

Si desea crear un nuevo plugin, el repositorio de [plantillas](https://github.com/terraform-linters/tflint-ruleset-template) se encuentra disponible para satisfacer esa especificación. Puede crear su propio repositorio haciendo clic sobre el botón "Use this template" y añadir nuevas reglas fácilmente basadas en otras reglas de referencia.

El repositorio de plantillas utiliza el [SDK](https://github.com/terraform-linters/tflint-plugin-sdk) que engloba `go-plugin` para su comunicación con TFLint. Vea también la sección de [arquitectura](https://github.com/terraform-linters/tflint-plugin-sdk#architecture) para obtener mas información de la arquitectura del sistema de plugins.

## 1. Creación de un repositorio a partir de la plantilla

Diríjase al repositorio de codigo [tflint-ruleset-template](https://github.com/terraform-linters/tflint-ruleset-template) y haga clic en el botón "Use this template". El nombre del repositorio debe ser `tflint-ruleset-*`.

## 2. Construir e instalar el plugin

El repositorio creado puede ser instalado localmente con mediante el comando `make install`. Habilite el plugin de la siguiente tal y como se muestra a continuación y verifique que el plugin instalado funciona correctamente.

```hcl
plugin "template" {
    enabled = true
}
```

```console
$ make install
go build
mkdir -p ~/.tflint.d/plugins
mv ./tflint-ruleset-template ~/.tflint.d/plugins
$ tflint -v
TFLint version 0.28.1
+ ruleset.template (0.1.0)
```

## 3. Cambiar o añadir las reglas

Cambie el nombre del conjunto de reglas y añada o edite las reglas. Después de hacer los cambios, puede comprobar el comportamiento con el comando `make install`. Vea también la referencia del API de [tflint-plugin-sdk](https://pkg.go.dev/github.com/terraform-linters/tflint-plugin-sdk) para la comunicación con entre los distintos procesos.

## 4. Crear una nueva versión en GitHub Release

Puede construir e instalar localmente su propio conjunto de reglas tal y como se ha descrito con anterioridad, pero también puede instalarlo automáticamente con `tflint --init`.

Los requisitos para proporcionar una instalación automática son los siguientes:

- Se debe publicar el plugin construido en GitHub Release.
- Se debe etiquetar la versión con un nombre similar `v1.1.1`.
- La versión debe contener un activo con el nombre similar a `tflint-ruleset-{nombre}_{GOOS}_{GOARCH}.zip`
- El archivo zip debe contener un archivo binario `tflint-ruleset-{nombre}` (`tflint-ruleset-{nombre}.exe` en Windows).
- La versión debe contener un archivo de verificación de sumas para el archivo zip con el nombre `checksums.txt`
- El archivo de comprobación de sumas debe contener un "hash" sha256 y un nombre de archivo.

Al firmar una nueva versión, ésta debe cumplir además con los siguientes requisitos:

- La versión debe contener un archivo de firma para el archivo de comprobación de sumas. El nombre del archivo debe ser `checksums.txt.sig`.
- El archivo de firma debe estar en formato binario OpenPGP.

Puede crear versiones que cumplan con estos requisitos de una forma fácil y sencilla, la configuración de GoReleaser en el repositorio de plantillas.
