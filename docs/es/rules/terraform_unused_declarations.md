# terraform_unused_declarations

No está permitida la declaración de variables, orígenes de datos o locales que nunca se utilizan.

## Ejemplo

```hcl
variable "not_used" {}

variable "used" {}
output "out" {
  value = var.used
}
```

```
$ tflint
1 issue(s) found:

Warning: variable "not_used" is declared but not used (terraform_unused_declarations)

  on config.tf line 1:
   1: variable "not_used" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.5/docs/rules/terraform_unused_declarations.md
 
```

## Porqué

Terraform ignorará las variables y los valores locales que no se utilicen. También refrescará los orígenes de datos declaradas independientemente de su utilización. Sin embargo, las variables no referenciadas probablemente sean indicativo de un error (y deben ser referenciadas) o de código eliminado (y deben ser eliminadas).

## Cómo solucionar el problema

Elimine su declaración. Para los bloques de código `variable` y `data`, elimine el bloque completo. Para un valor `local`, elimine el atributo delo bloque de código `locals`.

Si bien eliminar las fuentes de datos no debería ser un problema, en general, y provocar efectos no deseados, tenga precaución al eliminarlas. Por ejemplo, si elimina el bloque `data "http"` hará que Terraform no realice una solicitud HTTP `GET` durante cada plan de ejecución. Si un origen de datos se utiliza, añada una anotación para ignorarla:

```tf
# tflint-ignore: terraform_unused_declarations
data "http" "example" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"
}
```
