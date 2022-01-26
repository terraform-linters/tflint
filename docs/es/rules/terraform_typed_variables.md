# terraform_typed_variables

No está permitida la utilización de declaraciones de tipo `variable` sin la definición del atributo `type`.

## Ejemplo

```hcl
variable "no_type" {
  default = "value"
}

variable "enabled" {
  default     = false
  description = "This is description"
  type        = bool
}
```

```
$ tflint
1 issue(s) found:

Warning: `no_type` variable has no type (terraform_typed_variables)

  on template.tf line 1:
   1: variable "no_type" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_typed_variables.md
 
```

## Porqué

Dado que el atributo `type` es opcional, no siempre es necesario declararlo. Pero esta regla es útil si se quiere forzar la declaración de un tipo.

## Cómo solucionar el problema

Añada el atributo `type` a la declaración de la varaible. Puede ver https://www.terraform.io/docs/configuration/variables.html#type-constraint para obtener más información sobre los tipos de datos.
