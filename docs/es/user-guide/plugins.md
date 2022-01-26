# Configurar plugins

Puede ampliar la funcionalidad de TFLint instalando cualquier plugin. Declare los plugins que desee utilizar en el archivo de configuración de la siguiente forma:

```hcl
plugin "foo" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/org/tflint-ruleset-foo"

  signing_key = <<-KEY
  -----BEGIN PGP PUBLIC KEY BLOCK-----

  mQINBFzpPOMBEADOat4P4z0jvXaYdhfy+UcGivb2XYgGSPQycTgeW1YuGLYdfrwz
  9okJj9pMMWgt/HpW8WrJOLv7fGecFT3eIVGDOzyT8j2GIRJdXjv8ZbZIn1Q+1V72
  AkqlyThflWOZf8GFrOw+UAR1OASzR00EDxC9BqWtW5YZYfwFUQnmhxU+9Cd92e6i
  ...
  KEY
}
```

Despues de declarar los atributos `version` y `source`, `tflint --init` instalará automáticamente los plugins indicados.

```console
$ tflint --init
Installing `foo` plugin...
Installed `foo` (source: github.com/org/tflint-ruleset-foo, version: 0.1.0)
$ tflint -v
TFLint version 0.28.1
+ ruleset.foo (0.1.0)
```

Vea también [Configurar TFLint](config.md) para obtener información sobre el esquema del archivo de configuración.

## Attributos

Esta sección describe los atributos reservados por TFLint. A excepción de estos, cada plugin puede extender el esquema definiendo cualquier atributo/bloque. Consulte la documentación de cada plugin para obtener más detalles.

### `enabled` (requerido)

Habilita el plugin. Si lo establece su valor a `false`, no se utilizarán las reglas aunque el plugin esté instalado.

### `source`

La URL de origen para instalar el plugin. Debe tener el formato `github.com/org/repositorio`.

### `version`

Versión del plugin. No es necesario anteponer el prefijo "v". Si se establece el atributos `source` no es necesario utilizar este parámetro. Las restricciones sobre las versiones (como `>= 0.3`) no están soportadas.

### `signing_key`

Clave de firma pública PGP del desarrollador del plugin. Cuando se establece este atributo, TFLint verificará automáticamente la firma del archivo de verificación descargado de GitHub. Se recomienda configurarlo para evitar posibles ataques.

Los plugins de la organización terraform-linters (AWS/GCP/Azure ruleset plugins) pueden utilizar la clave de firma incorporada, por lo que se puede omitir este atributo.

## Directorio de plugins

Los plugins se suelen instalar en el directorio `~/.tflint.d/plugins`. De manera execpcioopnal, si ya tiene un directorio `./.tflint.d/plugins` en su directorio de trabajo, los plugins se pueden instalar ahí.

Los plugins instalados automáticamente se colocan como `[plugin dir]/[source]/[version]/tflint-ruleset-[name]`. (`tflint-ruleset-[name].exe` en Windows).

Si desea cambiar el directorio por defecto de los plugins, puede hacer con el atributo [`plugin_dir`](config.md#plugin_dir) en el archivo de configuración o mediante la variable de entorno `TFLINT_PLUGIN_DIR`.

## Evitar limitaciones

Cuando instala los plugins mediante la ejecución del comando `tflint --init`, se realizar llamadas al API de GitHub para obtener los metadatos de las versiones. Normalmente se trata de una solicitud no autentificada con un límite de 60 solicitudes por hora.

https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting

Esta limitación puede ser un problema si necesita ejecutar `--init` frecuentemente, como por ejemplo en entornos de CI. Si quiere incrementar ese límite de peticiones, puede enviar una solicitud autenticada estableciendo un token de acceso OAuth2 en la variable de entorno `GITHUB_TOKEN`.

También es una buena idea almacenar en caché el directorio de plugins, ya que TFLint sólo enviará peticiones si los plugins no están instalados. Vea también la sección [ejmplo de configuración de tflint](https://github.com/terraform-linters/setup-tflint#usage).

## Instalación manual

Tambiente puede instalar los plugins manualmente. Esta opcion es útil, principalmente, para el desarrollo de plugins o para los plugins que no están publicados en GitHub. En ese caso, omita los atributos `source` y `version`.

```hcl
plugin "foo" {
  enabled = true
}
```

Cuando se activa el plugin, TFLint invoca el binario `tflint-ruleset-[name]` (`tflint-ruleset-[name].exe` en Windows) en el directorio de plugins (Por ejemplo, `~/.tflint.d/plugins/tflint-ruleset-[name]`). Por lo tanto, debe mover el binario en al directorio con anterioridad.
