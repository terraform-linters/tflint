# terraform_naming_convention

Fuerza la utilización de convenciones de nomenclatura para los siguientes bloques:

* Recursos.
* Variables de emtrada.
* Declaraciones de tipo `output`.
* Declaraciones de tipo `local`.
* Módulos.
* Orígenes de datos.

## Configuración

Nombre | Por defecto | Valor
--- | --- | ---
enabled | `false` | Booleano
format | `snake_case` | `snake_case`, `mixed_snake_case`, `none` o un formato personalizado definido mediante el atributo `custom_formats`.
custom | `""` | Representación en forma de cadena de una expresión regular de golang con la que debe coincidir el nombre del bloque.
custom_formats | `{}` | Definición de formatos personalizados que se pueden utilizar en el atributo `format`.
data | | Configuración del bloque para sobreescribir la convención de nomenclatura de los orígenes de datos.
locals | | Configuración del bloque para sobreescribir la convención de nomenclatura de los valores locales.
module | | Configuración del bloque para sobreescribir la convención de nomenclatura de los módulos.
output | | Configuración del bloque para sobreescribir la convención de nomenclatura de los `output`
resource | | Configuración del bloque para sobreescribir la convención de nomenclatura de los recursos.
variable | | Configuración del bloque para sobreescribir la convención de nomenclatura de las variables de entrada.


#### `format`

La opción `format` define los formatos permitidos para las etiquetas de los bloques de código de Terraform.

Esta opción acepta uno de los siguientes valores:

* `snake_case` - formato snake_case estandar - todos los caracteres deben estar en minúsculas, y se permite la utilización de guiones bajos.
* `mixed_snake_case` - formato snake_case modificado - todos los caracteres pueden estar en mayúsculas o minúsculas y se permite la utilización de guiones bajos.
* `none` - significa que "no se comprobará el formato de este bloque". Esta opción resulta de utilidad si no se desea imponer ningún formato particular para un bloque de código.

#### `custom`

La opción `custom` define una regex personalizada que debe coincidir el identificador. Esta opción le permite tener un control más granular sobre los identificadores, permitiéndole forzar ciertos patrones y cadenas.

#### `custom_formats`

El atributo `custom_formats` define formatos adicionales que se pueden utilizar con la opción `format` option. Al igual que la opción `custom`, le permite definir una expresión regular personalizada que debe coincidir con el identificador, también le permite proporcionar una descripción que se mostrará cuando la comprobación falle. Tambien, le permite reutilizar una expresión regular personalizada.

Este atributo es un mapa, donde las claves son los identificadores de los formatos personalizados, y los valores son objetos con claves `regex` y `description`.

## Ejemplo

### Por defecto: Aplica snake_case para todos los bloques

#### Configuración de la regla

```hcl
rule "terraform_naming_convention" {
  enabled = true
}
```

#### Ejemplo de código fuente de Terraform

```hcl
data "aws_eip" "camelCase" {
}

data "aws_eip" "valid_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: data name `camelCase` must match the following format: snake_case (terraform_naming_convention)

  on template.tf line 1:
   1: data "aws_eip" "camelCase" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```


### Expresiones de nomenclatura personalizada para todos los bloques

#### Configuración de la regla

```hcl
rule "terraform_naming_convention" {
  enabled = true

  custom = "^[a-zA-Z]+([_-][a-zA-Z]+)*$"
}
```

#### Ejemplo de código fuente de Terraform

```hcl
resource "aws_eip" "Invalid_Name_With_Number123" {
}

resource "aws_eip" "Name-With_Dash" {
}
```

```
$ tflint
1 issue(s) found:

Notice: resource name `Invalid_Name_With_Number123` must match the following RegExp: ^[a-zA-Z]+([_-][a-zA-Z]+)*$ (terraform_naming_convention)

  on template.tf line 1:
   1: resource "aws_eip" "Invalid_Name_With_Number123" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
```


### Custom format for all blocks

#### Configuración de las reglas

```hcl
rule "terraform_naming_convention" {
  enabled = true
  format = "custom_format"

  custom_formats = {
    custom_format = {
      description = "Custom Format"
      regex       = "^[a-zA-Z]+([_-][a-zA-Z]+)*$"
    }
  }
}
```

#### Ejemplo de código fuente de Terraform

```hcl
resource "aws_eip" "Invalid_Name_With_Number123" {
}

resource "aws_eip" "Name-With_Dash" {
}
```

```
$ tflint
1 issue(s) found:

Notice: resource name `Invalid_Name_With_Number123` must match the following format: Custom Format (terraform_naming_convention)

  on template.tf line 1:
   1: resource "aws_eip" "Invalid_Name_With_Number123" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```


### Sobreescribre la configuración por defecto para un tipo de bloque específico 

#### Configuración de la regla

```hcl
rule "terraform_naming_convention" {
  enabled = true

  module {
    custom = "^[a-zA-Z]+(_[a-zA-Z]+)*$"
  }
}
```

#### Ejemplo de código fuente de Terraform

```hcl
// data name enforced with default snake_case
data "aws_eip" "eip_1a" {
}

module "valid_module" {
  source = ""
}

module "invalid_module_with_number_1a" {
  source = ""
}
```

```
$ tflint
1 issue(s) found:

Notice: module name `invalid_module_with_number_1a` must match the following RegExp: ^[a-zA-Z]+(_[a-zA-Z]+)*$ (terraform_naming_convention)

  on template.tf line 9:
   9: module "invalid_module_with_number_1a" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```

### Deshabilitada para un tipo de bloque específico

#### Configuración de la regla

```hcl
rule "terraform_naming_convention" {
  enabled = true

  module {
    format = "none"
  }
}
```

#### Ejemplo de código fuente de Terraform

```hcl
// data name enforced with default snake_case
data "aws_eip" "eip_1a" {
}

// module names will not be enforced
module "Valid_Name-Not-Enforced" {
  source = ""
}
```


### Deshabilitada para todos los bloques, habilitada para un tipo de bloque específico

#### Configuración de la regla

```hcl
rule "terraform_naming_convention" {
  enabled = true
  format  = "none"

  local {
    format = "snake_case"
  }
}
```

#### Ejemplo de código fuente de Terraform

```hcl
// Data block name not enforced
data "aws_eip" "EIP_1a" {
}

// Resource block name not enforced
resource "aws_eip" "EIP_1b" {
}

// local variable names enforced
locals {
  valid_name   = "valid"
  invalid-name = "dashes are not allowed with snake_case"
}
```

```
$ tflint
1 issue(s) found:

Notice: local value name `invalid-name` must match the following format: snake_case (terraform_naming_convention)

  on template.tf line 12:
  12: invalid-name = "dashes are not allowed with snake_case"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_naming_convention.md
 
```

## Porqué

La utilización de convenciones de nombres es opcional, por lo que no es necesario seguirlas.
Pero esta regla es útil si quiere forzar la autilización de convenciones de nomenclatura que se indican en [Buenas prácticas para la nomenclatura de plugins con Terraform](https://www.terraform.io/docs/extend/best-practices/naming.html).

## Cómo solucionar el problema

Actualice la etiqueta del bloque según el formato o la expresión regular personalizada.
