# Ideas Futuras para mdev

## Estado Actual
**Versión:** 1.0.0  
**Fecha:** 2026-03-01  
**Changes SDD Completados:** 15  
**Tests:** 96+ pasando  
**Plataformas:** Linux, macOS, Windows

---

## Ideas de Mejoras

### 1. Soporte para Monorepos
**Utilidad:** Alta  
**Descripción:** Detectar y manejar proyectos Nx, Turborepo, Rush  
**Comandos propuestos:**
- `mdev monorepo detect` - Detectar tipo de monorepo
- `mdev monorepo graph` - Mostrar dependencias entre apps
- `mdev monorepo build` - Build ordenado por dependencias
- `mdev monorepo clean` - Limpiar caches de todos los proyectos

**Motivación:** Cada vez más equipos usan monorepos. Actualmente mdev trabaja con un solo proyecto a la vez.

---

### 2. Sistema de Plugins
**Utilidad:** Media-Alta  
**Descripción:** Permitir plugins de terceros para frameworks específicos  
**Ejemplos:**
- Plugin para Ionic
- Plugin para NativeScript
- Plugin para Capacitor
- Plugin para herramientas internas de la empresa

**Motivación:** No podemos soportar todos los frameworks. Plugins permiten extensibilidad sin modificar el core.

---

### 3. Integración con Docker
**Utilidad:** Media  
**Descripción:** Detectar y configurar entornos Docker para mobile dev  
**Comandos propuestos:**
- `mdev docker setup` - Crear Dockerfile para el proyecto
- `mdev docker build` - Build usando Docker
- `mdev docker shell` - Entrar a shell del contenedor

**Motivación:** Algunos equipos prefieren Docker para consistencia total del entorno.

---

### 4. Team Sync / Configuración Compartida
**Utilidad:** Alta  
**Descripción:** Sincronizar configuración entre miembros del equipo  
**Comandos propuestos:**
- `mdev team init` - Crear configuración de equipo
- `mdev team sync` - Descargar config del equipo
- `mdev team share` - Compartir mi config con el equipo

**Motivación:** Todos los devs de un equipo deberían tener el mismo entorno. Evita "a mí me funciona".

---

### 5. Soporte para CI/CD Avanzado
**Utilidad:** Media  
**Descripción:** Mejor integración con GitHub Actions, GitLab CI, etc.  
**Features:**
- `mdev ci setup` - Crear workflow de CI
- `mdev ci validate` - Validar configuración de CI
- `mdev ci run` - Ejecutar CI localmente

**Motivación:** Facilitar la adopción de buenas prácticas de CI/CD en mobile.

---

### 6. Análisis de Performance
**Utilidad:** Media  
**Descripción:** Detectar cuellos de botella en builds  
**Comandos propuestos:**
- `mdev perf analyze` - Analizar tiempos de build
- `mdev perf compare` - Comparar builds
- `mdev perf suggest` - Sugerir optimizaciones

**Motivación:** Builds lentos = devs frustrados = menos productividad.

---

### 7. Gestión de Certificados (iOS)
**Utilidad:** Alta (para iOS)  
**Descripción:** Manejo de certificados y provisioning profiles  
**Comandos propuestos:**
- `mdev ios certs list` - Listar certificados
- `mdev ios certs validate` - Validar certificados
- `mdev ios certs sync` - Sincronizar con Apple Developer

**Motivación:** Configuración de iOS es un dolor de cabeza. Automatizarlo ahorra horas.

---

### 8. Backup y Restore de Entorno
**Utilidad:** Media  
**Descripción:** Backup completo del entorno de desarrollo  
**Comandos propuestos:**
- `mdev backup create` - Backup de todo el entorno
- `mdev backup restore` - Restaurar desde backup
- `mdev backup list` - Listar backups

**Motivación:** Cambiar de máquina o recuperarse de un crash sin perder tiempo.

---

### 9. Integración con IDEs
**Utilidad:** Media  
**Descripción:** Plugins para VS Code, IntelliJ, etc.  
**Features:**
- Botón "Run mdev doctor" en la IDE
- Highlight de problemas en tiempo real
- Quick fixes desde la IDE

**Motivación:** Los devs pasan todo el día en la IDE. Integración = adopción.

---

### 10. Analytics y Telemetría (Opt-in)
**Utilidad:** Baja (para usuarios), Alta (para mantainers)  
**Descripción:** Recolectar estadísticas de uso (anónimas)  
**Datos:**
- Qué comandos se usan más
- Qué errores son más comunes
- Tiempo de onboarding

**Motivación:** Mejorar el producto basado en datos reales de uso.

---

## Prioridad de Implementación

| Prioridad | Feature | Esfuerzo | Impacto |
|-----------|---------|----------|---------|
| 1 | Team Sync | Medio | Alto |
| 2 | Soporte Monorepos | Alto | Alto |
| 3 | iOS Certificates | Medio | Alto |
| 4 | Sistema de Plugins | Alto | Medio |
| 5 | CI/CD Avanzado | Medio | Medio |
| 6 | Performance Analysis | Medio | Medio |
| 7 | Docker Integration | Alto | Medio |
| 8 | Backup/Restore | Bajo | Medio |
| 9 | IDE Plugins | Alto | Medio |
| 10 | Analytics | Bajo | Bajo |

---

## Notas de Implementación

### Para Team Sync
- Podría usar un archivo `.mdev-team.yaml` en el repo
- O un servidor central (más complejo)
- Inicialmente: archivo en repo es suficiente

### Para Monorepos
- Detectar `nx.json`, `turbo.json`, `rush.json`
- Leer `workspace.json` o similar
- Mostrar grafo de dependencias

### Para Plugins
- Definir interfaz: `Plugin interface { Name() string, Check() error, Commands() []Command }`
- Cargar desde `~/.mdev/plugins/`
- O desde el PATH: `mdev-plugin-*`

---

## Recursos Útiles

- GitHub Releases API: https://docs.github.com/en/rest/releases
- Go Plugin System: https://github.com/hashicorp/go-plugin
- Nx Graph: https://nx.dev/core-features/explore-graph
- Docker SDK for Go: https://docs.docker.com/engine/api/sdk/

---

## Fecha de Creación
2026-03-01

## Autor
Lucas Videla

## Licencia
MIT
