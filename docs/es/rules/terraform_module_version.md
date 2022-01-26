# terraform_module_version

No está permitida la utilización de módulos de Terraform descargados desde [Terraform Registry](https://www.terraform.io/docs/language/modules/sources.html#terraform-registry) sin especificar una versión concreta.

## Configuración

Nombre | Descripción | Defecto | Tipo
--- | --- | --- | ---
exact | Requiere una versión exacta | `false` | Booleano

```hcl
rule "terraform_module_version" {
  enabled = true
  exact = false # default
}
```

## Ejemplo

```tf
module "exact" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "1.0.0"
}

module "range" {
  source  = "terraform-aws-modules/vpc/aws"
  version = ">= 1.0.0"
}

module "latest" {
  source  = "terraform-aws-modules/vpc/aws"
}
```

```
$ tflint
1 issue(s) found:

Warning: module "latest" should specify a version (terraform_module_version)

  on main.tf line 11:
  11: module "latest" {

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_module_version.md
```

### Parámetro exact

```hcl
rule "terraform_module_version" {
  enabled = true
  exact = true
}
```

```tf
module "exact" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "1.0.0"
}

module "range" {
  source  = "terraform-aws-modules/vpc/aws"
  version = ">= 1.0.0"
}
```

```
$ tflint
1 issue(s) found:

Warning: module "range" should specify an exact version, but a range was found (terraform_module_version)

  on main.tf line 8:
   8:   version = ">= 1.0.0"

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_module_version.md
```

## Porqué

En la [documentación sobre la versión de los módulos](https://www.terraform.io/docs/language/modules/syntax.html#version) de Terraform se afirma:

> Cuando se utilizan módulos instalados desde un registro de módulos, se recomienda restringir explícitamente los números de versión aceptables para evitar cambios inesperados o no deseados.

Cuando no se especifica el parámetro `version`, Terraform descargará la última versión disponible en el registro. El uso de una nueva versión mayor de un módulo podría causar la destrucción de los recursos existentes, o la creación de nuevos recursos que no son compatibles con las versiones anteriores. Por lo general, debería limitar los módulos a una versión principal específica.

### Versiones exactas

Dependiendo de su flujo de trabajo, es posible que quiera obligar que los módulos especifiquen una versión  _exacta_ por el parámetro `exact = true` para esta regla. Al hacer esto no permitirá que un módulo incluya múltiples restricciones de versión separadas por comas, o cualquier [operador de restricción](https://www.terraform.io/docs/language/expressions/version-constraints.html#version-constraint-syntax) que no sea `=`. Las versiones exactas se utilizan a menudo con gestores de dependencia automatizados como [Dependabot](https://docs.github.com/en/code-security/supply-chain-security/keeping-your-dependencies-updated-automatically/about-dependabot-version-updates) y [Renovate](https://docs.renovatebot.com), que propondrá automáticamente un pull request para actualizar el módulo cuando se publique una nueva versión.

Tenga en cuenta que el módulo puede incluir otros módulos hijos, que tienen sus propias restricciones sobre la versión. TFLint
_no_ comprueba las restricciones de versión establecidas en los módulos hijos **la activación de esta regla no garantiza que `terraform init` sea determinista**. Utilice los [archivos de bloqueo de dependencias de Terraform](https://www.terraform.io/docs/language/dependency-lock.html) para garantizar que Terraform utilice siempre la misma versión de todos los módulos (y proveedores) hasta que los actualice explícitamente

## Cómo solucionar el problema

Especifique el parametro `version`. Si se especifica el parámetro `exact = true`, el valor introducido debe ser una versión exacta.
