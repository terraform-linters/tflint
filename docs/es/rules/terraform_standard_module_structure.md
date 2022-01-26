# terraform_standard_module_structure

Fuera la utilización de módulos que cumplan con las normas de [estructura estándar de los módulos](https://www.terraform.io/docs/modules/index.html#standard-module-structure) de Terraform.

## Ejemplo

_main.tf_
```hcl
variable "v" {}
```

```
$ tflint
1 issue(s) found:

Warning: variable "v" should be moved from main.tf to variables.tf (terraform_standard_module_structure)

  on main.tf line 1:
   1: variable "v" {}

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.0/docs/rules/terraform_standard_module_structure.md
```

## Porqué

La documentación de Terraform describe una [estructura estándar para los módulos](https://www.terraform.io/docs/modules/structure.html). Como mínimo un módulo debe tener los archivos `main.tf`, `variables.tf` y `outputs.tf`. Los bloques de tipo `variable` o `output` deben incluirse en sus correspondientes archivos.

## Cómo solucionar el problema

* Mover los bloques de código a los archivos convencionales según sea necesario.
* Cree archivos vacíos aunque no haya bloques de tipo `variable` o `output` definidos.
