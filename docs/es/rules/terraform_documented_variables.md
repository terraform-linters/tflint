# terraform_documented_variables

No está permitida la utilización de declaraciones de tipo `variable` sin descripción.

## Ejemplo

```hcl
variable "no_description" {
  default = "value"
}

variable "empty_description" {
  default = "value"
  description = ""
}

variable "description" {
  default = "value"
  description = "This is description"
}
```

```
$ tflint
2 issue(s) found:

Notice: `no_description` variable has no description (terraform_documented_variables)

  on template.tf line 1:
   1: variable "no_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_variables.md

Notice: `empty_description` variable has no description (terraform_documented_variables)

  on template.tf line 5:
   5: variable "empty_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_variables.md
 
```

## Porqué

Cómo el atributo `description` es un valor opcional, no siempre es necesario escribirlo. Esta regla puede resultar de utilidad si quiere obligar a los desarrolladores a escribir una descripción, siendo especialmente utili si se combina con [terraform-docs](https://github.com/segmentio/terraform-docs).

## Cómo solucionar el problema

En la definición de las variables, escriba una descripción que no sea una cadena vacía.
