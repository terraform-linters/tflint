# terraform_required_version

No está permitida la utilización de bloques de configuración de tipo `terraform` sin el atributo `require_version` definido.

## Configuración

```hcl
rule "terraform_required_version" {
  enabled = true
}
```

## Ejemplo

```hcl
terraform {
  required_version = ">= 1.0" 
}
```

```
$ tflint
1 issue(s) found:

Warning: terraform "required_version" attribute is required

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_required_version.md 
```

## Porqué

El atributo `required_version` se puede utilizar para restringir las versiones de Terraform CLI que se pueden utilizar con su configuración.
Si la versión que está ejecutando de Terraform no coincide con las restricciones especificadas, Terraform producirá un error y saldrá sin realizar ninguna acción.

## Cómo solucionar el problema

Añada el atributo `required_version` al bloque de configuración de Terraform.
