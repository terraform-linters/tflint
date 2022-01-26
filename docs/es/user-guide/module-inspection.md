# Inspección de modulos

Por defecto, TFLint sólo inspecciona el modulo principal. Opcionalmente, tambien puede inspeccionar las[llamadas a módulos](https://www.terraform.io/docs/configuration/blocks/modules/syntax.html#calling-a-child-module). Cuando esta opción está habilitada, TFLint analiza cada llamada (por ejem,plo, los bloques `module`) y emite cualquier incidencia detectada a partir de las variables de entrada especificadas. La inspección de módulos está diseñada para que se ejecute en todas las llamas a los módulos, donde el `source` sea local (`./*`) o remoto (registry, git, etc.). 

```hcl
module "aws_instance" {
  source        = "./module"

  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"
}
```

```console
$ tflint --module
1 issue(s) found:

Error: instance_type is not a valid value (aws_instance_invalid_type)

  on template.tf line 5:
   5:   instance_type = "t1.2xlarge"

Callers:
   template.tf:5,19-31
   module/instance.tf:5,19-36

```

## Advertencias

* El modo de inspección de modulos _no busca recursivamente_ los módulos de Terraform. Busca los bloques `module` en el módulo principal desde donde se ha invocado a TFLint.
* Las incidencias _deben_ estar asociada a una variable que se haya pasado al módulo. Si se detecta una incidencia dentro de un módulo hijo en una expresión que no hace referencia a una variable (`var`), esta incidencia será descartada.
* Las reglas que evalúan la sintaxis en lugar del contenido _deben_ ignorar los módulos hijos. 

Si quiere evaluar todas las reglas de TFLint en módulos distintos al módulo principal, debe indicar su ruta directamente a TFLint.

## Habilitar la inspección de modulos

La inspección de módulos está desactivada por defecto y puede activarse con el indicador `--module`. Tambien se puede habilitar en el archivo de configuración de TFLint:

```hcl
config {
  module = true
}
```

Debe ejecutar `terraform init` antes de llamar a TFLint con la inspección de modulos, de esta forma los modulos se cargan en el directorio `.terraform`. Puede utilizar el indicador `--ignore-module` si desea omitir la inspección de un modulo en concreto:

```
tflint --ignore-module=./module
```
