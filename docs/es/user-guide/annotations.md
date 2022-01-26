# Anotaciones

Los comentarios de anotación permiten desactivar las reglas en líneas específicas:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

En este ejemplo, la anotación sólo funciona para la misma línea o para la línea inferior. Puede utilizar `tflint-ignore: all` si desea ignorar todas las reglas.
