# terraform_deprecated_index

No está permitida la utilización de la antigua sintaxis para obtener los elementos de una lista mediante un índice, utilizando puntos.

## Ejemplo

```hcl
locals {
  list  = ["a", "b", "c"]
  value = list.0 
}
```

```
$ tflint
1 issue(s) found:

Warning: List items should be accessed using square brackets (terraform_deprecated_index)

  on example.tf line 3:
   3:   value = list.0

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.1/docs/rules/terraform_deprecated_index.md
```

## Porqué

La versión v0.12 de Terraform soporta la utilización de corchetes para acceder mediante un índice a los elementos de una lista. Sin embargo, para mantener la compatibilidad con la versión v0.11, Terraform continua soportando el acceso a los elementos de una lista mediante la antigua sintaxis de puntos que se utiliza normalmente para acceder a los atributos. Aunque Terraform no muestra ninguna  advertencia cuando se utilizar esta sintaxis, ya no está documentada y se desaconseja su utilización.

## Cómo solucionar el problema

Cambie a la nueva sintaxis de corchetes cuando acceda a los elementos de una lista, incluyendo los recursos que utilizan `count`.
