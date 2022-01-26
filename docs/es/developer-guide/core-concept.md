# Concepto básicos

TFLint es sólo un contenedor (wrapper) de Terraform. La carga de la configuración, la evaluación de las expresiones, etc., dependen del API interno de Terraform, y sólo proporciona una interfaz para hacerlas como una herramienta linter.

A excepción de algunas reglas, todas Las reglas son proporcionadas por los plugins. Técnicamente, los plugins se lanzan como otros procesos, se comunican vía RPC, y reciben los resultados de la inspección del proceso del plugin.

A continuación se describen los componentes más importantes para entender su comportamiento:

- `tflint`
- Este componente es el núcleo de TFLint y funciona como un contenedor (wrapper) de Terraform. Permite el acceso a `terraform/configs.Config` y `terraform/terraform.BuiltinEvalContext`, etc.
- `plugin`
- Este componente proporciona el sistema de plugins de TFLint. Incluye el sistema de para localizar los plugins y una implementación de servidor que responde a las peticiones de los plugins.
- `cmd`
- Este paquete es el punto de entrada de la aplicación.
