# terraform_documented_outputs

No está permitida la utilización de declaraciones de tipo `output` sin descripción.

## Ejemplo

```hcl
output "no_description" {
  value = "value"
}

output "empty_description" {
  value = "value"
  description = ""
}

output "description" {
  value = "value"
  description = "This is description"
}
```

```
$ tflint
2 issue(s) found:

Notice: `no_description` output has no description (terraform_documented_outputs)

  on template.tf line 1:
   1: output "no_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_outputs.md

Notice: `empty_description` output has no description (terraform_documented_outputs)

  on template.tf line 5:
   5: output "empty_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_outputs.md
 
```

## Porqué

Cómo el atributo `description` es un valor opcional, no siempre es necesario escribirlo. Esta regla puede resultar de utilidad si quiere obligar a los desarrolladores a escribir una descripción, siendo especialmente utili si se combina con [terraform-docs](https://github.com/segmentio/terraform-docs).

## Cómo solucionar el problema

En la definición de los `outputs`, escriba una descripción que no sea una cadena vacía.
