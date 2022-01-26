# terraform_module_pinned_source

No está permitida la utilización de un repositorio de Git o de Mercurial como origen de un módulo sin forzar la utilización de una versión específica.

## Configuración

Nombre | Por defecto | Valor
--- | --- | ---
enabled | true | Booleano
style | `flexible` | `flexible`, `semver`
default_branches | `["master", "main", "default", "develop"]` | 

```hcl
rule "terraform_module_pinned_source" {
  enabled = true
  style = "flexible"
  default_branches = ["dev"]
}
```

Al establecer el atributo, `default_branches` las ramas configuradas se añadirán a las predeterminadas en lugar de sobreescribirlas.

## Ejemplo

### style = "flexible"

In the "flexible" style, all sources must be pinned to non-default version.
Al utilizar el estilo de "flexible", todos los recursos se deben fijar a una versión diferente a la versión por defecto.

```hcl
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}

module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}

module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}

module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=feature"
}
```

```
$ tflint
3 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_pinned_source)

  on template.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=master" uses a default branch as ref (master) (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=master"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "hg::http://hashicorp.com/consul.hg?rev=default" uses a default branch as rev (default) (terraform_module_pinned_source)

  on template.tf line 10:
  10:   source = "hg::http://hashicorp.com/consul.hg?rev=default"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

```

### style = "semver"

Al utilizar el estilo de "semver", todos los recursos se deben fijar a la versión semántica de referencia. Esta opción es más estricta que el estilo "flexible".


```hcl
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}

module "pinned_to_branch" {
  source = "git://hashicorp.com/consul.git?ref=feature"
}

module "pinned_to_version" {
  source = "git://hashicorp.com/consul.git?ref=v1.2.0"
}
```

```
$ tflint
2 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_pinned_source)

  on template.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=feature" uses a ref which is not a semantic version string (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=feature"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

```

## Porqué

Terraform le permite obtener módulos de los repositorios de control de fuentes. Si no fija la versión que desea utilizar, la dependencia que necesita puede introducir cambios que rompan la ejecución de Terraform de manera inesperada.

Para evitarlo, especifique siempre una versión explícita de sus dependencias.

Fijar a una referencia mutable, como una rama, sigue permitiendo intencionados que rompan la ejecución de Terraform. Utilizar Semver puede ayudarle a evitar este problema.

## Cómo solucionar el problema

Especifique una versión que le ayude a fijar sus dependencias. Si utiliza repositorios de Git, no debe utilizar la rama "master". Si utiliza repositorios de Mercurial, no debe utilizar la rama "default".

En el estilo de "semver": especifique una versión semántica para fijar las dependecias, esta versión debe tener la forma `vX.Y.Z`. La utilización de la `v` inicial, es opcional.
