# terraform_unused_required_providers

Fuerza la comprobación de que todos los `required_providers` declarados, se utilizan en el modulo.

## Configuración

```hcl
rule "terraform_unused_required_providers" {
  enabled = true
}
```

## Ejemplos

```hcl
terraform {
  required_providers {
    null = {
      source = "hashicorp/null"
    }
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: provider 'null' is declared in required_providers but not used by the module (terraform_unused_required_providers)

  on main.tf line 3:
   3:     null = {
   4:       source = "hashicorp/null"
   5:     }

Reference: https://github.com/terraform-linters/tflint/blob/v0.22.0/docs/rules/terraform_unused_required_providers.md
```

## Porqué

El bloque `required_providers` debe especificar los proveedores utilizados directamente por el módulo Terraform. Terraform descargará todos los proveedores especificados durante la ejecución del comando `terraform init`. Si se eliminan todos los recursos de un determinado proveedor pero mantiene el bloque `required_providers`, Terraform continuará descargando el proveedor.

En general, cada módulo debe indicar su propio bloque `required_providers` para cada proveedor utilizado. Terraform recorrerá el gráfico de módulos y encontrará una versión adecuada para todos los proveedores, o bien mostrará un error si los módulos requieren versiones potencialmente conflictivas.

## Cómo solucionar el problema

Si ya no utiliza el proveedor, puede elimonar el bloque `required_providers`.

Si el proveedor se utiliza en uno o más módulos hijos pero no se utiliza directamente desde el módulo desde el cúal se invocó TFLint, corte y pegue el bloque `required_providers` en esos módulos.

Si el proveedor se utiliza en uno o más módulos hijos y prefiere utilizar una única definición del bloque `required_providers`, puede ignorar la advertencia:

```tf
terraform {
  required_providers {
    # tflint-ignore: terraform_unused_required_providers
    null = {
      source = "hashicorp/null"
    }
  }
}
```

Esto afectará a su capacidad para ejecutar `terraform` directamente en el móodulo hijo, especialmente si se utilizan proveedores externos al espacio de nombres predeterminado `hashicorp` o especifíca una versión en concreto  mediante el parámetro `version` dentro del bloque `required_providers` ([recomendación](./terraform_required_providers.md)).
