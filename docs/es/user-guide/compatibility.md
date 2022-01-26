# Compatibilidad con Terraform
TFLint agrupa los paquetes internos de Terraform como una libreria. Esto permite que el lenguaje de Terraform sea analizado correctamente incluso si el binario de Terraform no está instalado en tiempo de ejecución.

Por otro lado, la semántica del lenguaje depende del comportamiento de una versión concreta del paquete. Por ejemplo, una configuración analizada por Terraform v1.0 puede ser analizada por el analizador de lenguaje v1.1. La versión actual es la v1.1.0.

La recomendación es hacer coincidir la versión de Terraform incluida en TFLint con la versión que realmente está utilizando. Sin embargo, el lenguaje Terraform garantiza cierta compatibilidad con versiones anteriores, por lo que diferentes versiones pueden no causar problemas inmediatos. Sin embargo, tenga en cuenta que pueden producirse falsos positivos/negativos debido a esto.

## Variables de entrada

Al igual que Terraform, se soporta el uso de los indicadores `--var`,`--var-file`, variables de entorno (`TF_VAR_*`) y la carga automática de las definiciones de las variables mediante los archivos (`terraform.tfvars` y `*.auto.tfvars`). Puede ver el artículo [Variables de entrada](https://www.terraform.io/docs/language/values/variables.html) para obtener más información.

Las variables de entrada se evalúan correctamente, al igual que Terraform:

```hcl
variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => "t2.micro"
}
```

## Valores con nombre

Actualmente, los [valores con nombre](https://www.terraform.io/docs/configuration/expressions/references.html) se soportan de manera parcial. Los siguientes valores con nombre están disponibles:

- `var.<NAME>`
- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`

Las expresiones que hacen referencia a los valores con nombre no incluidos en la lista anterior (por ejemplo, `locals.*`, `count.*`, `each.*`, etc.) son excluidos de la inspección de TFLint.

```hcl
locals {
  instance_family = "t2"
}

resource "aws_instance" "foo" {
  instance_type = "${local.instance_family}.micro" # => Not an error, it will be ignored because it marks as unknown
}
```

## Funciones integradas

Las [funciones integradas](https://www.terraform.io/docs/configuration/functions.html) son totalmente compatibles.

## Variables de entorno

Actualmente, se soportan las siguientes variables de entorno:

- [TF_VAR_name](https://www.terraform.io/docs/commands/environment-variables.html#tf_var_name)
- [TF_DATA_DIR](https://www.terraform.io/docs/commands/environment-variables.html#tf_data_dir)
- [TF_WORKSPACE](https://www.terraform.io/docs/commands/environment-variables.html#tf_workspace)
