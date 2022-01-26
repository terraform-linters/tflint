# Reglas

Las reglas se proporcionan a través de plugins con conjuntos de reglas, pero las reglas para el lenguaje Terraform están construidas dentro el binario TFLint. A continuación se muestra una lista de reglas disponibles.

|Regla|Descripción|Habilitada|
| --- | --- | --- |
|[terraform_comment_syntax](terraform_comment_syntax.md)|No está permitida la utilización de comentarios mediante `//`, en su lugar está permitido la utilización de comentarios mediante `#`.||
|[terraform_deprecated_index](terraform_deprecated_index.md)|No está permitida la utilización de la antigua sintaxis para obtener los elementos de una lista mediante un índice, utilizando puntos.||
|[terraform_deprecated_interpolation](terraform_deprecated_interpolation.md)|No está permitida la utilización de interpolación de variables al estilo de Terraform v0.11.|✔|
|[terraform_documented_outputs](terraform_documented_outputs.md)|No está permitida la utilización de declaraciones de tipo `output` sin descripción.||
|[terraform_documented_variables](terraform_documented_variables.md)|No está permitida la utilización de declaraciones de tipo `variable` sin descripción.||
|[terraform_module_pinned_source](terraform_module_pinned_source.md)|No está permitida la utilización de un repositorio de Git o de Mercurial como origen de un módulo sin forzar la utilización de una versión específica.|✔|
|[terraform_module_version](terraform_module_version.md)|No está permitida la utilización de módulos de Terraform descargados desde un registro sin especificar una versión concreta.|✔|
|[terraform_naming_convention](terraform_naming_convention.md)|Fuerza la utilización de convenciones de nomenclatura para los recursos, orígenes de datos, etc.||
|[terraform_required_providers](terraform_required_providers.md)|No está permitida la utilización de proveedores de Terraform sin restricciones de versión a través `required_providers`.||
|[terraform_required_version](terraform_required_version.md)|No está permitida la utilización de bloques de configuración de tipo `terraform` sin el atributo `require_version` definido.||
|[terraform_standard_module_structure](terraform_standard_module_structure.md)|Fuerza la utilización de módulos que cumplan con la estructura de módulos estándar de Terraform.||
|[terraform_typed_variables](terraform_typed_variables.md)|No está permitida la utilización de declaraciones de tipo `variable` sin la definición del atributo `type`.||
|[terraform_unused_declarations](terraform_unused_declarations.md)|No está permitida la declaración de variables, orígenes de datos o locales que nunca se utilizan.||
|[terraform_unused_required_providers](terraform_unused_required_providers.md)|Fuerza la comprobación de que todos los `required_providers` declarados, se utilizan en el modulo.||
|[terraform_workspace_remote](terraform_workspace_remote.md)| No está permitida la utilización de `terraform.workspace` cómo un backend "remoto" de ejecución remota.|✔|
