# Configurar TFLint

Puede cambiar el comportamiento de TFLint mediante los distintos indicadores del CLI y también utilizando archivos de configuración. Por defecto, TFLint busca los ficheros `.tflint.hcl` según la siguiente lista de prioridad:

- En el directorio actual (`./.tflint.hcl`)
- En el directorio de inicio de cada usuario (`~/.tflint.hcl`)

El fichero de configuración está escrito en lenguaje [HCL](https://github.com/hashicorp/hcl). A continuación se muestra un ejemplo:

```hcl
config {
  plugin_dir = "~/.tflint.d/plugins"

  module = true
  force = false
  disabled_by_default = false

  ignore_module = {
    "terraform-aws-modules/vpc/aws"            = true
    "terraform-aws-modules/security-group/aws" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]
  variables = ["foo=bar", "bar=[\"baz\"]"]
}

plugin "aws" {
  enabled = true
  version = "0.4.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

rule "aws_instance_invalid_type" {
  enabled = false
}
```

También puede utilizar otro archivo como archivo de configuración con el indicador `--config`:

```
$ tflint --config other_config.hcl
```

### `plugin_dir`

Estable el directorio de los plugins. Por defecto, este directorio se encuentra en `~/.tflint.d/plugins` (o `./.tflint.d/plugins`). Puede ver [configurar plugins](plugins.md#advanced-usage)

### `module`

Indicador del CLI: `--module`

Habilita la [inspección de módulos](module-inspection.md).

### `force`

Indicador del CLI: `--force`

Devuelve un código de salida nulo incluso si TFLint encuentra algún problema. TFLint devuelve por defecto los siguientes códigos de salida:

- 0: TFLint no ha encontrado incidencias.
- 1: Se han producido errores.
- 2: No se han producido errores, pero se han encontrado incidencias.

### `disabled_by_default`

Indicador del CLI: `--only`

Sólo se activan las reglas habilitadas específicamente en la configuración o mediante la línea de comandos. Las reglas restantes, incluidas las predeterminadas, están desactivadas. Nota, el uso del parámetro `--only` en la línea de comandos ingnorará otras reglas habilitadas mediante los parámetros `--enable-rule` o `--disable-rule`.

```hcl
config {
  disabled_by_default = true
  # other options here...
}

rule "aws_instance_invalid_type" {
  enabled = true
}

rule "aws_instance_previous_type" {
  enabled = true
}
```

```console
$ tflint --only aws_instance_invalid_type --only aws_instance_previous_type
```

### `ignore_module`

Indicador del CLI: `--ignore-module`

Omitir las inspecciones de las llamadas a los módulos en la [inpsección de módulos](module-inspection.md). Tenga en cuenta que debe especificar los orígenes de los módulos en vez de los identificadores de los mismos, para la compatibilidad con versiones anteriores.

```hcl
config {
  module = true
  ignore_module = {
    "terraform-aws-modules/vpc/aws"            = true
    "terraform-aws-modules/security-group/aws" = true
  }
}
```

```console
$ tflint --ignore-module terraform-aws-modules/vpc/aws --ignore-module terraform-aws-modules/security-group/aws
```

### `varfile`

Indicador del CLI: `--var-file`

Establezce las variables de Terraform desde los archivos `tfvars`. Si existe el archivo `terraform.tfvars` o existe alg'un fichero `*.auto.tfvars`, se cargarán automáticamente.

```hcl
config {
  varfile = ["example1.tfvars", "example2.tfvars"]
}
```

```console
$ tflint --var-file example1.tfvars --var-file example2.tfvars
```

### `variables`

Indicador del CLI: `--var`

Establece una variable de Terraform a partir de un valor pasado. Este indicador puede utilizarse varias veces.

```hcl
config {
  variables = ["foo=bar", "bar=[\"baz\"]"]
}
```

```console
$ tflint --var "foo=bar" --var "bar=[\"baz\"]"
```

### bloques `rule`

Indicador del CLI : `--enable-rule`, `--disable-rule`

Puede configurar las reglas de TFLint mediante bloques `rule`. La implementación de cada regla especifica si estará habilitada por defecto. En algunos conjuntos de reglas, la mayoría de las reglas están desactivadas por defecto. Utilice los bloques `rule` para habilitarlas:

```hcl
rule "terraform_unused_declarations" {
  enabled = true
}
```

El atributo `enabled` es necesario para todos los bloques `rule`. Para las reglas que están activadas por defecto, establezca `enabled = false` para desactivar la regla:

```hcl
rule "aws_instance_previous_type" {
  enabled = false
}
```

Algunas reglas permiten la utilización de atributos adicionales que modifican su comportamiento. Consulte la documentación de cada regla para obtener más información.

### bloques `plugin`

Puede declarar el plugin a utilizar. Puede ver [Configurar plugins](plugins.md) para obtener más información.
