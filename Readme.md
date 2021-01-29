Culpa
PkgGoDev goreportcard

El paquete de fallas proporciona middleware go http que facilita la inyección de fallas en su servicio. Use el paquete de fallas para rechazar solicitudes entrantes, responder con un error HTTP, inyectar latencia en un porcentaje de sus solicitudes o inyectar cualquiera de sus propias fallas personalizadas.

Caracteristicas
El paquete de errores funciona a través del middleware estándar go http . Primero crea un Injector, que es un middleware con el código que se ejecutará en la inyección. Luego envuelve eso Injectoren un Faultque maneja la lógica sobre cuándo ejecutar tu Injector.

Actualmente, existen tres tipos de inyectores: SlowInjector, ErrorInjector, y RejectInjector. Cada uno de estos inyectores se puede configurar a través de un Faultpara que se ejecute en un pequeño porcentaje de sus solicitudes. También puede configurar la lista Faultde bloqueo / lista de permisos para ciertas rutas.

Consulte la sección de uso a continuación para ver un ejemplo de cómo comenzar y el godoc para obtener más documentación.

Limitaciones
Este paquete es útil para probar escenarios de falla de forma segura en servicios go que pueden hacer uso de net/httpcontroladores / middleware.

Un escenario de falla común que no podemos simular perfectamente son las solicitudes descartadas. El RejectInjectorsiempre devolverá inmediatamente al usuario, sino que en muchos casos las solicitudes pueden ser dejados sin enviar nunca una respuesta. La mejor manera de simular este escenario usando el paquete de fallas es encadenar a SlowInjectorcon un tiempo de espera muy largo frente a un eventual RejectInjector.

Estado
Este proyecto se encuentra en un estado estable y con soporte. No hay planes para introducir nuevas funciones significativas, sin embargo, agradecemos y alentamos cualquier idea y contribución de la comunidad. Las contribuciones deben seguir las pautas de nuestro CONTRIBUTING.md .

Uso
// 
paquete main.go main

import (
         "net / http" 
        "tiempo"

        "github.com/github/go-fault"
)

var  mainHandler  =  http . HandlerFunc ( func ( w http. ResponseWriter , r  * http. Request ) {
         http . Error ( w , http . StatusText ( http . StatusOK ), http . StatusOK )
})

func  main () {
         slowInjector , _  : =  falla . NewSlowInjector ( tiempo . Segundo  *  2 )
         slowFault , _  : =  falla . NewFault ( slowInjector ,
                 falla . WithEnabled ( verdadero ),
                 falla . WithParticipation ( 0.25 ),
                 falla . WithPathBlocklist ([] string {"/ ping" , "/ salud" }),
        )

        // Agrega 2 segundos de latencia al 25% de nuestras solicitudes 
        handlerChain  : =  slowFault . Handler ( mainHandler )

        http . ListenAndServe ( "127.0.0.1:3000" , handlerChain )
}
Desarrollo
Este paquete utiliza herramientas estándar de go para pruebas y desarrollo. El idioma go es todo lo que necesita para contribuir. Las pruebas utilizan el popular testificar / afirmar que se descargará automáticamente la primera vez que ejecute las pruebas. Las acciones de GitHub también ejecutarán un linter usando golangci-lint después de presionar. También puede descargar el linter y usarlo golangci-lint runpara ejecutarlo localmente.

Pruebas
El paquete de fallas tiene pruebas extensas que se ejecutan en acciones de GitHub en cada inserción. La cobertura del código es del 100% y se publica como un artefacto en cada ejecución de Acciones.

También puede ejecutar pruebas localmente:

$ ir prueba -v -cover -race. / ...
[...]
PASAR
cobertura: 100.0% de declaraciones
ok github.com/github/go-fault 0.575s
Benchmarks
Es seguro dejar implementado el paquete de fallas incluso cuando no se está ejecutando una inyección de fallas. Si bien la falla está deshabilitada, hay una degradación del rendimiento insignificante en comparación con la eliminación del paquete de la ruta de solicitud. Mientras esté habilitado, puede haber pequeñas diferencias de rendimiento, pero este solo será el caso mientras ya esté inyectando fallas.

Se proporcionan puntos de referencia para comparar sin fallas, con fallas deshabilitadas y con fallas habilitadas. Los puntos de referencia se cargan como artefactos en Acciones de GitHub y puede descargarlos desde cualquier flujo de trabajo de validación .

También puede ejecutar evaluaciones comparativas localmente (salida de ejemplo):

$ ir prueba -run = XXX -bench =.
goos: darwin
goarch: amd64
paquete: github.com/github/go-fault
BenchmarkNoFault-8 684826 1734 ns / op
BenchmarkFaultDisabled-8 675291 1771 ns / op
BenchmarkFaultErrorZeroPercent-8 667903 1823 ns / op
BenchmarkFaultError100Percent-8 663661 1833 ns / op
PASAR
ok github.com/github/go-fault 8.814s
Mantenedores
@lingrino

Colaboradores
@mrfaizal @vroldanbet @fatih

Licencia
Este proyecto tiene la licencia MIT .
