# terraform_deprecated_interpolation

No está permitida la utilización de interpolación de variables al estilo de Terraform v0.11.

## Ejemplo

```hcl
resource "aws_instance" "deprecated" {
    instance_type = "${var.type}"
}

resource "aws_instance" "new" {
    instance_type = var.type
}
```

```
$ tflint
1 issue(s) found:

Warning: Interpolation-only expressions are deprecated in Terraform v0.12.14 (terraform_deprecated_interpolation)

  on example.tf line 2:
   2:     instance_type = "${var.type}"

Reference: https://github.com/terraform-linters/tflint/blob/v0.14.0/docs/rules/terraform_deprecated_interpolation.md
 
```

## Porqué

Terraform v0.12 introduce una nueva sintaxis para la interpolación de variables, pero sigue soportando la antigua sintaxis de interpolación de variables de la versión 0.11 de Terraform, por motivos de compatibilidad.

Terraform mostrará las advertencias de diagnóstico cuando se utilice la sintaxis de interpolación de variables obsoletas. En consonancia con su política de desaprobación, se mostrarán con un mensaje de error en la próxima versión (v0.13). Siguiendo la misma lógica TFLint mostrara un problema en lugar de una advertencia.

Terraform will currently print diagnostic warnings when deprecated interpolations are used. Consistent with its deprecation policy, they will raise errors in the next major release (v0.13). TFLint emits an issue instead of a warning with the same logic.

## Cómo solucionar el problema

Cambie a la nueva sintaxis de interpolación. Puede ver las notas de la versión de Terraform 0.12.14 para obtener más información: https://github.com/hashicorp/terraform/releases/tag/v0.12.14
