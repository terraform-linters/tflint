# terraform_required_providers

No está permitida la utilización de proveedores de Terraform sin restricciones de versión a través `required_providers`.

## Configuración

```hcl
rule "terraform_required_providers" {
  enabled = true
}
```

## Ejemplos

```hcl
provider "template" {}
```

```
$ tflint
1 issue(s) found:

Warning: Missing version constraint for provider "template" in "required_providers" (terraform_required_providers)

  on main.tf line 1:
   1: provider "template" {}

Reference: https://github.com/terraform-linters/tflint/blob/v0.18.0/docs/rules/terraform_required_providers.md
```

<hr>

```hcl
provider "template" {
  version = "2"
}
```

```
$ tflint
2 issue(s) found:

Warning: provider.template: version constraint should be specified via "required_providers" (terraform_required_providers)

  on main.tf line 1:
   1: provider "template" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.18.0/docs/rules/terraform_required_providers.md

Warning: Missing version constraint for provider "template" in "required_providers" (terraform_required_providers)

  on main.tf line 1:
   1: provider "template" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.18.0/docs/rules/terraform_required_providers.md
```

## Porqué

Los proveedores son complementos liberados a un ritmo separado del de el propio Terraform, por lo que tienen sus propios números de versión. Para su uso en producción, debe restringir las versiones aceptadas de los proveedores a través de sus configuraciones, para asegurar que las futuras versiones que potencialmente pueden introducir cambios que rompan la ejecución de Terraform, no se instalen automáticamente mediante la ejecución del comando `terraform init`. 

## How To Fix

Añada el bloque [`required_providers`](https://www.terraform.io/docs/configuration/terraform.html#specifying-required-provider-versions) al bloque de configruación `terraform` e incluya las versiones disponibles para todos los proveedores de Terraform. Por ejemplo:

```tf
terraform {
  required_providers {
    template = "~> 2.0"
  }
}
```

Las restricciones en la versión de un  proveedor se pueden especificar utilizando el [argumento version dentro de un bloque de tipo provider](https://www.terraform.io/docs/configuration/providers.html#provider-versions) para tener compatibilidad con versiones anteriores. Actualmente se desaconseja la utilización de este enfoque, especialmente cuando su código de Terraform utiliza módulos hijos.
