# terraform_comment_syntax

No está permitida la utilización de comentarios mediante `//`, en su        lugar está permitido la utilización de comentarios mediante `#`.

## Ejemplo

```hcl
# Bien
// Mal
```

```
$ tflint
1 issue(s) found:

Warning: Single line comments should begin with # (terraform_comment_syntax)

  on main.tf line 2:
   2: // Bad

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.0/docs/rules/terraform_typed_variables.md
```

## Porqué

El lenguaje Terraform admite dos sintaxis diferentes para los comentarios de una sola línea: `#` y `//`. Sin embargo, `#` es el estilo de comentario por defecto y debería utilizarse en la mayoría de los casos.

* [Sintaxis de configuración: Comentarios](https://www.terraform.io/docs/configuration/syntax.html#comments)

## Cómo solucionar el problema

Sustituya la doble barra inicial (`//`) en su comentarios con el signo de número (`#`).
