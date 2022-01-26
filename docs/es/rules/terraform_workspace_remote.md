# terraform_workspace_remote

No está permitida la utilización de `terraform.workspace` cómo un backend "remoto" de ejecución remota.

Si las operaciones remotas están [deshabilitadas](https://www.terraform.io/docs/cloud/run/index.html#disabling-remote-operations) para su espacio de trabajo, puede desactivar esta regla de forma segura:

```hcl
rule "terraform_workspace_remote" {
  enabled = false
}
```

## Ejemplo 

```hcl
terraform {
  backend "remote" {
    # ...
  }
}

resource "aws_instance" "a" {
  tags = {
    workspace = terraform.workspace
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: terraform.workspace should not be used with a 'remote' backend (terraform_workspace_remote)

  on example.tf line 8:
   8:   tags = {
   9:     workspace = terraform.workspace
  10:   }

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.5/docs/rules/terraform_workspace_remote.md
```

## Porqué

La configuración de Terraform puede incluir el nombre del [espacio de trabajo actual](https://www.terraform.io/docs/state/workspaces.html#current-workspace-interpolation) utilizando la secuencia de interpolación `${terraform.workspace}`. Sin embargo, cuando los espacios de trabajo de Terraform Cloud ejecutan Terraform de forma remota, la CLI de Terraform siempre utiliza el espacio de trabajo `por defecto`.

El backend [remoto](https://www.terraform.io/docs/backends/types/remote.html) se utiliza con los espacios de trabajo de Terraform Cloud. Incluso si se establece un `prefijo` en el bloque `workspaces`, este valor será ignorado durante las ejecuciones remotas.

Para obtener más información, puede ver la [documentación de los espacios de trabajo remotos del backend](https://www.terraform.io/docs/backends/types/remote.html#workspaces).

## Cómo solucionar el problema

Considere la posibilidad de añadir una variable a su configuración y establecerla en cada espacio de trabajo en la nube:

```tf
variable "workspace" {
  type        = string
  description = "The workspace name" 
}
```

También puede nombrar la variable basándose en lo que el sufijo del espacio de trabajo representa en su configuración (por ejemplo, entorno).
