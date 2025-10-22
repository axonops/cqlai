# CQLAI - Shell Moderno de Cassandra CQL

<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">
</div>

**CQLAI** es un terminal interactivo rápido y portátil para Cassandra (CQL), construido en Go. Proporciona una alternativa moderna y fácil de usar a `cqlsh` con una interfaz de terminal avanzada, análisis de comandos del lado del cliente y funciones de productividad mejoradas.

**Las funciones de IA son completamente opcionales** - CQLAI funciona perfectamente como un shell CQL independiente sin ninguna configuración de IA o claves API.

El comando cqlsh original está escrito en Python, lo que requiere que Python esté instalado en el sistema. cqlai está compilado en un único binario ejecutable, sin requerir dependencias externas. Este proyecto proporciona binarios para las siguientes plataformas:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


Está construido con [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), y [Lip Gloss](https://github.com/charmbracelet/lipgloss) para la hermosa interfaz de terminal. Un gran reconocimiento al equipo del driver gocql de Cassandra por implementar las últimas funcionalidades de Cassandra [gocql](https://github.com/apache/cassandra-gocql-driver)


---

## Estado del Proyecto

**CQLAI está listo para producción** y se utiliza activamente en entornos de desarrollo, pruebas y producción con clústeres de Cassandra. La herramienta proporciona una alternativa completa y estable a `cqlsh` con características y rendimiento mejorados.

### Lo que Funciona
- Todas las operaciones y consultas CQL principales
- Soporte completo de meta-comandos (`DESCRIBE`, `SHOW`, `CONSISTENCY`, etc.)
- Análisis de comandos del lado del cliente (ligero, sin dependencia de ANTLR)
- Importación/exportación de datos con `COPY TO/FROM` (formatos CSV y Parquet)
- Conexiones SSL/TLS y autenticación
- Tipos Definidos por el Usuario (UDTs) y tipos de datos complejos
- Modo batch para scripting y automatización
- Soporte de formato Apache Parquet para intercambio eficiente de datos
- Autocompletado con tabulador para palabras clave CQL, tablas, columnas y keyspaces
- Tamaño de binario pequeño (~43MB, 53% más pequeño que versiones anteriores)
- **Opcional**: Generación de consultas potenciada por IA (OpenAI, Anthropic, Gemini)

### Próximamente
- Conciencia de contexto de IA mejorada
- Servicio MCP de Cassandra
- Optimizaciones de rendimiento adicionales

Te animamos a **probar CQLAI hoy** y ayudar a dar forma a su desarrollo. Tu retroalimentación y contribuciones son invaluables para hacer de este el mejor shell CQL para la comunidad de Cassandra. Por favor [reporta problemas](https://github.com/axonops/cqlai/issues) o [contribuye](https://github.com/axonops/cqlai/pulls).

---

## Características

- **Shell CQL Interactivo:** Ejecuta cualquier consulta CQL que tu clúster de Cassandra soporte.
- **Interfaz de Terminal Enriquecida:**
    - Una aplicación de terminal de múltiples capas y pantalla completa con búfer de pantalla alternativo (preserva el historial del terminal).
    - Tabla virtualizada y desplazable para resultados con carga automática de datos, previniendo sobrecarga de memoria en consultas grandes.
    - Modos de navegación avanzados con atajos de teclado estilo vim.
    - Soporte completo de ratón incluyendo desplazamiento con rueda y selección de texto.
    - Barra de estado/pie de página fija mostrando detalles de conexión, latencia de consulta y estado de sesión (consistencia, trazado).
    - Superposiciones modales para historial, ayuda y autocompletado de comandos.
- **Soporte de Apache Parquet:**
    - Formato de datos columnar de alto rendimiento para flujos de trabajo de análisis y aprendizaje automático.
    - Exporta tablas de Cassandra a archivos Parquet con el comando `COPY TO`.
    - Importa archivos Parquet a Cassandra con inferencia automática de esquema.
    - Conjuntos de datos particionados con estructuras de directorios estilo Hive.
    - Columnas virtuales TimeUUID / timestamp para particionamiento inteligente basado en tiempo.
    - Soporte para todos los tipos de datos de Cassandra incluyendo UDTs, colecciones y vectores.
- **Generación de Consultas Potenciada por IA (Opcional):**
    - Conversión de lenguaje natural a CQL usando proveedores de IA (OpenAI, Anthropic, Gemini).
    - Generación de consultas con conciencia de esquema y contexto automático.
    - Vista previa segura y confirmación antes de la ejecución.
    - Soporte para operaciones complejas incluyendo DDL y DML.
    - **Requiere configuración de clave API** - no necesaria para la funcionalidad principal.
- **Configuración:**
    - Configuración simple mediante `cqlai.json` en el directorio actual o `~/.cqlai.json`.
    - Soporte para conexiones SSL/TLS con autenticación por certificado.
- **Binario Único:** Distribuido como un único binario estático sin dependencias externas. Inicio rápido y huella pequeña.

## Instalación

Puedes instalar `cqlai` de varias maneras. Para instrucciones detalladas incluyendo gestores de paquetes (APT, YUM) y Docker, consulta la [Guía de Instalación](docs/INSTALLATION.md).

### Binarios Precompilados

Descarga el binario apropiado para tu sistema operativo y arquitectura desde la página de [**Releases**](https://github.com/axonops/cqlai/releases).


### Usando Go

```bash
go install github.com/axonops/cqlai/cmd/cqlai@latest
```

### Desde el Código Fuente

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
go build -o cqlai cmd/cqlai/main.go
```

### Usando Docker

```bash
# Construir la imagen
docker build -t cqlai .

# Ejecutar el contenedor
docker run -it --rm --name cqlai-session cqlai --host tu-host-cassandra
```

## Uso

### Modo Interactivo

Conectar a un host de Cassandra:
```bash
# Con contraseña en línea de comandos (no recomendado - visible en ps)
cqlai --host 127.0.0.1 --port 9042 --username cassandra --password cassandra

# Con solicitud de contraseña (seguro - contraseña oculta)
cqlai --host 127.0.0.1 --port 9042 -u cassandra
# Password: [entrada oculta]

# Usando variable de entorno (seguro para scripts/contenedores)
export CQLAI_PASSWORD=cassandra
cqlai --host 127.0.0.1 -u cassandra
```

O usa un archivo de configuración:
```bash
# Crear configuración desde el ejemplo
cp cqlai.json.example cqlai.json
# Editar cqlai.json con tu configuración, luego ejecuta:
cqlai
```

### Opciones de Línea de Comandos

```bash
cqlai [opciones]
```

#### Opciones de Conexión
| Opción | Corta | Descripción |
|--------|-------|-------------|
| `--host <host>` | | Host de Cassandra (sobrescribe config) |
| `--port <puerto>` | | Puerto de Cassandra (sobrescribe config) |
| `--keyspace <keyspace>` | `-k` | Keyspace predeterminado (sobrescribe config) |
| `--username <usuario>` | `-u` | Usuario para autenticación |
| `--password <contraseña>` | `-p` | Contraseña para autenticación* |
| `--no-confirm` | | Desactivar confirmaciones |
| `--connect-timeout <segundos>` | | Tiempo de espera de conexión (predeterminado: 10) |
| `--request-timeout <segundos>` | | Tiempo de espera de petición (predeterminado: 10) |
| `--debug` | | Habilitar registro de depuración |

*\*Nota: La contraseña puede proporcionarse de tres maneras:*
1. *Línea de comandos con `-p` (no recomendado - visible en la lista de procesos)*
2. *Solicitud interactiva cuando se usa `-u` sin `-p` (recomendado)*
3. *Variable de entorno `CQLAI_PASSWORD` (bueno para automatización)*

#### Opciones de Modo Batch
| Opción | Corta | Descripción |
|--------|-------|-------------|
| `--execute <declaración>` | `-e` | Ejecutar declaración CQL y salir |
| `--file <archivo>` | `-f` | Ejecutar CQL desde archivo y salir |
| `--format <formato>` | | Formato de salida: ascii, json, csv, table |
| `--no-header` | | No mostrar encabezados de columna (CSV) |
| `--field-separator <sep>` | | Separador de campos para CSV (predeterminado: ,) |
| `--page-size <n>` | | Filas por lote (predeterminado: 100) |

#### Opciones Generales
| Opción | Corta | Descripción |
|--------|-------|-------------|
| `--config-file <ruta>` | | Ruta al archivo de configuración (sobrescribe ubicaciones predeterminadas) |
| `--help` | `-h` | Mostrar mensaje de ayuda |
| `--version` | `-v` | Imprimir versión y salir |

### Ejemplos de Modo Batch

Ejecutar declaraciones CQL de forma no interactiva (compatible con cqlsh):

```bash
# Ejecutar una sola declaración
cqlai -e "SELECT * FROM system_schema.keyspaces;"

# Ejecutar desde un archivo
cqlai -f script.cql

# Entrada por tubería
echo "SELECT * FROM users;" | cqlai

# Controlar formato de salida
cqlai -e "SELECT * FROM users;" --format json
cqlai -e "SELECT * FROM users;" --format csv --no-header

# Controlar tamaño de paginación
cqlai -e "SELECT * FROM large_table;" --page-size 50
```

### Comandos Básicos

- **Ejecutar CQL:** Escribe cualquier declaración CQL y presiona Enter.
- **Meta-Comandos:**
  ```sql
  DESCRIBE KEYSPACES;
  USE mi_keyspace;
  DESCRIBE TABLES;
  CONSISTENCY QUORUM;
  TRACING ON;
  PAGING 50;
  EXPAND ON;  -- Modo de salida vertical
  SOURCE 'script.cql';  -- Ejecutar script CQL
  ```
- **Generación de Consultas Potenciada por IA:**
  ```sql
  .ai ¿Qué keyspaces hay?
  .ai ¿Qué columnas tiene la tabla users?
  .ai crear una tabla para almacenar inventario de productos
  .ai eliminar pedidos de más de 1 año de la tabla orders
  ```

### Atajos de Teclado

#### Navegación y Control
| Atajo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `↑`/`↓` | Navegar historial de comandos | Igual |
| `Ctrl+P`/`Ctrl+N` | Anterior/Siguiente en historial de comandos | Igual |
| `Alt+N` | Mover a siguiente línea en historial | `Option+N` |
| `Tab` | Autocompletar comandos y nombres de tablas/keyspaces | Igual |
| `Ctrl+C` | Limpiar entrada / Cancelar paginación / Cancelar operación (dos veces para salir) | `⌘+C` o `Ctrl+C` |
| `Ctrl+D` | Salir de la aplicación | `⌘+D` o `Ctrl+D` |
| `Ctrl+R` | Buscar en historial de comandos | `⌘+R` o `Ctrl+R` |
| `Esc` | Activar/desactivar modo de navegación / Cancelar paginación / Cerrar modales | Igual |
| `Enter` | Ejecutar comando / Cargar siguiente página (durante paginación) | Igual |

#### Edición de Texto
| Atajo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `Ctrl+A` | Saltar al inicio de la línea | Igual |
| `Ctrl+E` | Saltar al final de la línea | Igual |
| `Ctrl+Izq`/`Ctrl+Der` | Saltar por palabra (o 20 caracteres) | Igual |
| `PgUp`/`PgDn` (en entrada) | Página izq/der en consultas largas | `Fn+↑`/`Fn+↓` |
| `Ctrl+K` | Cortar desde el cursor hasta el final de la línea | Igual |
| `Ctrl+U` | Cortar desde el inicio hasta el cursor | Igual |
| `Ctrl+W` | Cortar palabra hacia atrás | Igual |
| `Alt+D` | Eliminar palabra hacia adelante | `Option+D` |
| `Ctrl+Y` | Pegar texto cortado previamente | Igual |

#### Cambio de Vista
| Atajo | Acción |
|----------|--------|
| `F2` | Cambiar a vista de consulta/historial |
| `F3` | Cambiar a vista de tabla |
| `F4` | Cambiar a vista de trazas (cuando el trazado está habilitado) |
| `F5` | Cambiar a vista de conversación IA |
| `F6` | Activar/desactivar tipos de datos de columna en encabezados de tabla |

#### Desplazamiento y Navegación de Tabla
| Atajo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `PgUp`/`PgDn` | Desplazar vista por página / Cargar más datos cuando esté disponible | `Fn+↑`/`Fn+↓` |
| `Espacio` | Cargar siguiente página cuando haya más datos disponibles | Igual |
| `Enter` (entrada vacía) | Cargar siguiente página cuando haya más datos disponibles | Igual |
| `Alt+↑`/`Alt+↓` | Desplazar vista por una sola fila (respeta límites de fila) | `Option+↑`/`Option+↓` |
| `Alt+←`/`Alt+→` | Desplazar tabla horizontalmente (tablas anchas) | `Option+←`/`Option+→` |
| `↑`/`↓` | Navegar filas de tabla (cuando está en modo navegación) | Igual |

#### Modo de Navegación (Vistas de Tabla/Trazas)
Presiona `Esc` para activar/desactivar el modo de navegación cuando veas tablas o trazas.

| Atajo | Acción en Modo de Navegación |
|----------|---------------------------|
| `j` / `k` | Desplazar abajo/arriba por una sola línea |
| `d` / `u` | Desplazar abajo/arriba por media página |
| `g` / `G` | Saltar al inicio/final de resultados |
| `<` / `>` | Desplazar izq/der por 10 columnas |
| `{` / `}` | Desplazar izq/der por 50 columnas |
| `0` / `$` | Saltar a primera/última columna |
| `Esc` | Salir del modo de navegación / Cancelar paginación si está activa |

#### Soporte de Ratón
| Acción | Función |
|--------|----------|
| Rueda del Ratón | Desplazamiento vertical con carga automática de datos |
| Alt+Rueda del Ratón | Desplazamiento horizontal en tablas |
| Shift+Rueda del Ratón | Desplazamiento horizontal (alternativa) |
| Ctrl+Rueda del Ratón | Desplazamiento horizontal (alternativa) |
| Shift+Clic+Arrastre | Seleccionar texto para copiar |
| Ctrl+Shift+C | Copiar texto seleccionado al portapapeles |
| Clic del Medio | Pegar desde el búfer de selección (Linux/Unix) |

**Nota para Usuarios de macOS:**
- La mayoría de atajos `Ctrl` funcionan tal cual en macOS, pero también puedes usar la tecla `⌘` (Comando) como alternativa
- La tecla `Alt` está etiquetada como `Option` en los teclados Mac
- Las teclas de función (F1-F6) pueden requerir mantener presionada la tecla `Fn` dependiendo de tu configuración de Mac

### Autocompletado con Tabulador

CQLAI proporciona autocompletado inteligente y consciente del contexto para acelerar tu flujo de trabajo. Presiona `Tab` en cualquier momento para ver las opciones de autocompletado disponibles.

#### Qué se Puede Autocompletar

**Palabras Clave y Comandos CQL:**
- Todas las palabras clave CQL: `SELECT`, `INSERT`, `CREATE`, `ALTER`, `DROP`, etc.
- Meta-comandos: `DESCRIBE`, `CONSISTENCY`, `COPY`, `SHOW`, etc.
- Tipos de datos: `TEXT`, `INT`, `UUID`, `TIMESTAMP`, etc.
- Niveles de consistencia: `ONE`, `QUORUM`, `ALL`, `LOCAL_QUORUM`, etc.

**Objetos de Esquema:**
- Nombres de keyspaces
- Nombres de tablas (dentro del keyspace actual)
- Nombres de columnas (cuando el contexto lo permite)
- Nombres de tipos definidos por el usuario
- Nombres de funciones y agregados
- Nombres de índices

**Autocompletados Conscientes del Contexto:**
```sql
-- Después de SELECT, sugiere nombres de columnas y palabras clave
SELECT <Tab>           -- Muestra: *, nombres de columnas, DISTINCT, JSON, etc.

-- Después de FROM, sugiere nombres de tablas
SELECT * FROM <Tab>    -- Muestra: tablas disponibles en el keyspace actual

-- Después de USE, sugiere nombres de keyspaces
USE <Tab>              -- Muestra: keyspaces disponibles

-- Después de DESCRIBE, sugiere tipos de objetos
DESCRIBE <Tab>         -- Muestra: KEYSPACE, TABLE, TYPE, etc.

-- Después del comando de consistencia
CONSISTENCY <Tab>      -- Muestra: ONE, QUORUM, ALL, etc.
```

**Autocompletado de Rutas de Archivo:**
```sql
-- Para comandos que aceptan rutas de archivo
SOURCE '<Tab>          -- Muestra: archivos en el directorio actual
SOURCE '/ruta/<Tab>    -- Muestra: archivos en /ruta/
```

#### Comportamiento del Autocompletado

- **Insensible a Mayúsculas:** Escribe `sel<Tab>` para obtener `SELECT`
- **Coincidencia Parcial:** Escribe parte de una palabra y presiona Tab
- **Múltiples Coincidencias:** Cuando hay múltiples opciones de autocompletado disponibles:
  - Primer Tab: Muestra autocompletado en línea si es único
  - Segundo Tab: Muestra todas las opciones disponibles en un modal
- **Filtrado Inteligente:** Los autocompletados se filtran según el contexto actual
- **Escape para Cancelar:** Presiona `Esc` para cerrar el modal de autocompletado

#### Ejemplos

```sql
-- Autocompletar nombre de tabla
SELECT * FROM us<Tab>
-- Completa a: SELECT * FROM users

-- Autocompletar nivel de consistencia
CONSISTENCY LOC<Tab>
-- Muestra: LOCAL_ONE, LOCAL_QUORUM, LOCAL_SERIAL

-- Autocompletar nombres de columnas después de SELECT
SELECT id, na<Tab> FROM users
-- Completa a: SELECT id, name FROM users

-- Autocompletar rutas de archivo para comando SOURCE
SOURCE 'sche<Tab>
-- Completa a: SOURCE 'schema.cql'

-- Autocompletar opciones del comando COPY
COPY users TO 'file.csv' WITH <Tab>
-- Muestra: HEADER, DELIMITER, NULLVAL, PAGESIZE, etc.

-- Mostrar todas las tablas cuando existen múltiples
SELECT * FROM <Tab>
-- Muestra modal con: users, orders, products, etc.
```

#### Consejos para Uso Efectivo

1. **Usa Tab liberalmente:** El sistema de autocompletado es inteligente y consciente del contexto
2. **Escribe caracteres mínimos:** A menudo 2-3 caracteres son suficientes para obtener un autocompletado único
3. **Usa para descubrir:** Presiona Tab en entrada vacía para ver qué está disponible
4. **Rutas de archivo:** Recuerda incluir comillas para autocompletado de rutas de archivo
5. **Navega autocompletados:** Usa las teclas de flecha para seleccionar entre múltiples opciones

## Comandos Disponibles

CQLAI soporta todos los comandos CQL estándar además de meta-comandos adicionales para funcionalidad mejorada.

### Comandos CQL
Ejecuta cualquier declaración CQL válida soportada por tu clúster de Cassandra:
- DDL: `CREATE`, `ALTER`, `DROP` (KEYSPACE, TABLE, INDEX, etc.)
- DML: `SELECT`, `INSERT`, `UPDATE`, `DELETE`
- DCL: `GRANT`, `REVOKE`
- Otros: `USE`, `TRUNCATE`, `BEGIN BATCH`, etc.

### Meta-Comandos

Los meta-comandos proporcionan funcionalidad adicional más allá del CQL estándar:

#### Gestión de Sesión
- **CONSISTENCY** `<nivel>` - Establecer nivel de consistencia (ONE, QUORUM, ALL, etc.)
  ```sql
  CONSISTENCY QUORUM
  CONSISTENCY LOCAL_ONE
  ```

- **PAGING** `<tamaño>` | OFF - Establecer tamaño de paginación de resultados
  ```sql
  PAGING 1000
  PAGING OFF
  ```

- **TRACING** ON | OFF - Habilitar/deshabilitar trazado de consultas
  ```sql
  TRACING ON
  SELECT * FROM users;
  TRACING OFF
  ```

- **OUTPUT** [FORMATO] - Establecer formato de salida
  ```sql
  OUTPUT          -- Mostrar formato actual
  OUTPUT TABLE    -- Formato de tabla (predeterminado)
  OUTPUT JSON     -- Formato JSON
  OUTPUT EXPAND   -- Formato vertical expandido
  OUTPUT ASCII    -- Formato de tabla ASCII
  ```

#### Descripción de Esquema
- **DESCRIBE** - Mostrar información de esquema
  ```sql
  DESCRIBE KEYSPACES                    -- Listar todos los keyspaces
  DESCRIBE KEYSPACE <nombre>            -- Mostrar definición de keyspace
  DESCRIBE TABLES                       -- Listar tablas en el keyspace actual
  DESCRIBE TABLE <nombre>               -- Mostrar estructura de tabla
  DESCRIBE TYPES                        -- Listar tipos definidos por el usuario
  DESCRIBE TYPE <nombre>                -- Mostrar definición de UDT
  DESCRIBE FUNCTIONS                    -- Listar funciones de usuario
  DESCRIBE FUNCTION <nombre>            -- Mostrar definición de función
  DESCRIBE AGGREGATES                   -- Listar agregados de usuario
  DESCRIBE AGGREGATE <nombre>           -- Mostrar definición de agregado
  DESCRIBE MATERIALIZED VIEWS           -- Listar vistas materializadas
  DESCRIBE MATERIALIZED VIEW <nombre>   -- Mostrar definición de vista
  DESCRIBE INDEX <nombre>               -- Mostrar definición de índice
  DESCRIBE CLUSTER                      -- Mostrar información del clúster
  DESC <keyspace>.<tabla>               -- Atajo para descripción de tabla
  ```

#### Exportación/Importación de Datos
- **COPY TO** - Exportar datos de tabla a archivo CSV o Parquet
  ```sql
  -- Exportación básica a CSV
  COPY users TO 'users.csv'

  -- Exportar a formato Parquet (autodetectado por extensión)
  COPY users TO 'users.parquet'

  -- Exportar a Parquet con formato y compresión explícitos
  COPY users TO 'data.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY'

  -- Exportar columnas específicas
  COPY users (id, name, email) TO 'users_partial.csv'

  -- Exportar con opciones
  COPY users TO 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|'

  -- Exportar a stdout
  COPY users TO STDOUT WITH HEADER = TRUE

  -- Opciones disponibles:
  -- FORMAT = 'CSV'/'PARQUET' -- Formato de salida (predeterminado: CSV, autodetectado)
  -- HEADER = TRUE/FALSE      -- Incluir encabezados de columna (solo CSV)
  -- DELIMITER = ','          -- Delimitador de campos (solo CSV)
  -- NULLVAL = 'NULL'        -- Cadena a usar para valores NULL
  -- PAGESIZE = 1000         -- Filas por página para exportaciones grandes
  -- COMPRESSION = 'SNAPPY'  -- Para Parquet: SNAPPY, GZIP, ZSTD, LZ4, NONE
  -- CHUNKSIZE = 10000       -- Filas por fragmento para Parquet
  ```

- **COPY FROM** - Importar datos CSV o Parquet a tabla
  ```sql
  -- Importación básica desde archivo CSV
  COPY users FROM 'users.csv'

  -- Importar desde archivo Parquet (autodetectado)
  COPY users FROM 'users.parquet'

  -- Importar desde Parquet con formato explícito
  COPY users FROM 'data.parquet' WITH FORMAT='PARQUET'

  -- Importar con fila de encabezado (CSV)
  COPY users FROM 'users.csv' WITH HEADER = TRUE

  -- Importar columnas específicas
  COPY users (id, name, email) FROM 'users_partial.csv'

  -- Importar desde stdin
  COPY users FROM STDIN

  -- Importar con opciones personalizadas
  COPY users FROM 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|' AND NULLVAL = 'N/A'

  -- Opciones disponibles:
  -- HEADER = TRUE/FALSE      -- Primera fila contiene nombres de columnas
  -- DELIMITER = ','          -- Delimitador de campos
  -- NULLVAL = 'NULL'        -- Cadena representando valores NULL
  -- MAXROWS = -1            -- Máximo de filas a importar (-1 = ilimitado)
  -- SKIPROWS = 0            -- Número de filas iniciales a saltar
  -- MAXPARSEERRORS = -1     -- Máximo de errores de análisis permitidos (-1 = ilimitado)
  -- MAXINSERTERRORS = 1000  -- Máximo de errores de inserción permitidos
  -- MAXBATCHSIZE = 20       -- Máximo de filas por inserción batch
  -- MINBATCHSIZE = 2        -- Mínimo de filas por inserción batch
  -- CHUNKSIZE = 5000        -- Filas entre actualizaciones de progreso
  -- ENCODING = 'UTF8'       -- Codificación del archivo
  -- QUOTE = '"'             -- Carácter de comillas para cadenas
  ```

- **CAPTURE** - Capturar salida de consulta a archivo (grabación continua)
  ```sql
  CAPTURE 'output.txt'          -- Comenzar a capturar a archivo de texto
  CAPTURE JSON 'output.json'    -- Capturar como JSON
  CAPTURE CSV 'output.csv'      -- Capturar como CSV
  SELECT * FROM users;
  CAPTURE OFF                   -- Detener captura
  ```

- **SAVE** - Guardar resultados de consulta mostrados a archivo (sin re-ejecutar)
  ```sql
  -- Primero ejecuta una consulta
  SELECT * FROM users WHERE status = 'active';

  -- Luego guarda los resultados mostrados en varios formatos:
  SAVE                           -- Diálogo interactivo (elegir formato y nombre de archivo)
  SAVE 'users.csv'               -- Guardar a CSV (formato autodetectado)
  SAVE 'users.json'              -- Guardar a JSON (formato autodetectado)
  SAVE 'users.txt' ASCII         -- Guardar como tabla ASCII
  SAVE 'data.csv' CSV            -- Especificar formato explícitamente

  -- Diferencias clave con CAPTURE:
  -- - SAVE exporta los resultados mostrados actualmente
  -- - No necesita re-ejecutar la consulta
  -- - Preserva los datos exactos mostrados en el terminal
  -- - Funciona con resultados paginados (guarda solo páginas cargadas)
  ```

#### Visualización de Información
- **SHOW** - Mostrar información de sesión
  ```sql
  SHOW VERSION          -- Mostrar versión de Cassandra
  SHOW HOST            -- Mostrar detalles de conexión actual
  SHOW SESSION         -- Mostrar toda la configuración de sesión
  ```

- **EXPAND** ON | OFF - Activar/desactivar modo de salida expandida
  ```sql
  EXPAND ON            -- Salida vertical (un campo por línea)
  SELECT * FROM users WHERE id = 1;
  EXPAND OFF           -- Salida de tabla normal
  ```

#### Ejecución de Scripts
- **SOURCE** - Ejecutar scripts CQL desde archivo
  ```sql
  SOURCE 'schema.cql'           -- Ejecutar script
  SOURCE '/ruta/a/script.cql'   -- Ruta absoluta
  ```

#### Ayuda
- **HELP** - Mostrar ayuda de comandos
  ```sql
  HELP                 -- Mostrar todos los comandos
  HELP DESCRIBE        -- Ayuda para comando específico
  HELP CONSISTENCY     -- Ayuda para niveles de consistencia
  ```

### Comandos de IA
- **.ai** `<consulta en lenguaje natural>` - Generar CQL desde lenguaje natural
  ```sql
  .ai mostrar todos los usuarios con estado activo
  .ai crear una tabla para almacenar sesiones de usuario
  .ai encontrar pedidos realizados en los últimos 30 días
  ```

## Configuración

CQLAI soporta múltiples métodos de configuración para máxima flexibilidad y compatibilidad con configuraciones existentes de Cassandra.

### Precedencia de Configuración

Las fuentes de configuración se cargan en el siguiente orden (las fuentes posteriores sobrescriben las anteriores):

1. **Archivos CQLSHRC** (para compatibilidad con configuraciones cqlsh existentes)
   - `~/.cassandra/cqlshrc` (ubicación estándar)
   - `~/.cqlshrc` (ubicación alternativa)
   - `$CQLSH_RC` (si se establece la variable de entorno)

2. **Archivos de configuración JSON de CQLAI**
   - `./cqlai.json` (directorio actual)
   - `~/.cqlai.json` (directorio home del usuario)
   - `~/.config/cqlai/config.json` (directorio de configuración XDG)

3. **Variables de entorno**
   - `CQLAI_HOST`, `CQLAI_PORT`, `CQLAI_KEYSPACE`, etc.
   - `CASSANDRA_HOST`, `CASSANDRA_PORT` (para compatibilidad)

4. **Banderas de línea de comandos** (prioridad más alta)
   - `--host`, `--port`, `--keyspace`, `--username`, `--password`, etc.

### Compatibilidad con CQLSHRC

CQLAI puede leer archivos CQLSHRC estándar usados por la herramienta tradicional `cqlsh`, haciendo la migración transparente.

**Secciones CQLSHRC soportadas:**
- `[connection]` - hostname, port, configuración ssl
- `[authentication]` - keyspace, ruta de archivo de credenciales
- `[auth_provider]` - módulo de autenticación y nombre de usuario
- `[ssl]` - configuración de certificados SSL/TLS

**Ejemplo de archivo CQLSHRC:**
```ini
; ~/.cassandra/cqlshrc
[connection]
hostname = cassandra.example.com
port = 9042
ssl = true

[authentication]
keyspace = mi_keyspace
credentials = ~/.cassandra/credentials

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client-key.pem
usercert = ~/certs/client-cert.pem
validate = true
```

Consulta [CQLSHRC_SUPPORT.md](docs/CQLSHRC_SUPPORT.md) para detalles completos de compatibilidad con CQLSHRC.

### Configuración JSON de CQLAI

Para características avanzadas y configuración de IA, CQLAI usa su propio formato JSON:

**Ejemplo `cqlai.json`:**
```json
{
  "host": "127.0.0.1",
  "port": 9042,
  "keyspace": "",
  "username": "cassandra",
  "password": "cassandra",
  "requireConfirmation": true,
  "consistency": "LOCAL_ONE",
  "pageSize": 100,
  "maxMemoryMB": 10,
  "connectTimeout": 10,
  "requestTimeout": 10,
  "debug": false,
  "historyFile": "~/.cqlai/history",
  "aiHistoryFile": "~/.cqlai/ai_history",
  "ssl": {
    "enabled": false,
    "certPath": "/ruta/a/client-cert.pem",
    "keyPath": "/ruta/a/client-key.pem",
    "caPath": "/ruta/a/ca-cert.pem",
    "hostVerification": true,
    "insecureSkipVerify": false
  },
  "ai": {
    "provider": "openai",
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

**Nota:** También puedes usar el campo `url` para sobrescribir el endpoint de la API para APIs compatibles con OpenAI:
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "tu-clave-api",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

### Configuración de Proveedor de IA

**Nota:** Las características de IA son completamente opcionales. CQLAI funciona como un shell CQL completo sin ninguna configuración de IA.

Para habilitar la generación de consultas potenciada por IA, configura tu proveedor preferido en la sección `ai` de tu archivo `cqlai.json`.

#### OpenAI (GPT-4 y GPT-3.5)

Usa OpenAI para generación de consultas de alta calidad y propósito general. Requiere una clave API de OpenAI.

- **Obtener Clave API:** [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **Modelos Recomendados:**
  - `gpt-4-turbo-preview` (predeterminado, recomendado para mejores resultados)
  - `gpt-3.5-turbo` (más rápido, más económico)

**Configuración:**
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

#### Anthropic (Claude 3)

Usa Anthropic para modelos potentes y conscientes del contexto. Ideal para consultas complejas y razonamiento. Requiere una clave API de Anthropic.

- **Obtener Clave API:** [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)
- **Modelos Recomendados:**
  - `claude-3-opus-20240229` (más potente)
  - `claude-3-sonnet-20240229` (predeterminado, rendimiento equilibrado)
  - `claude-3-haiku-20240307` (más rápido)

**Configuración:**
```json
{
  "ai": {
    "provider": "anthropic",
    "apiKey": "sk-ant-...",
    "model": "claude-3-sonnet-20240229"
  }
}
```

#### Google Gemini

Usa Google Gemini para un modelo rápido y capaz de Google. Requiere una clave API de Google AI Studio.

- **Obtener Clave API:** [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)
- **Modelo Recomendado:**
  - `gemini-pro` (predeterminado)

**Configuración:**
```json
{
  "ai": {
    "provider": "gemini",
    "apiKey": "...",
    "model": "gemini-pro"
  }
}
```

#### Ollama (Modelos Locales)

Usa Ollama para ejecutar modelos de IA localmente o conectarte a APIs compatibles con OpenAI. Ollama te permite ejecutar modelos de lenguaje potentes en tu propio hardware sin enviar datos a servicios externos.

- **Comenzar:** [ollama.ai](https://ollama.ai)
- **Modelos Recomendados:**
  - `llama3.2` (Llama 3.2 de Meta)
  - `codellama` (Llama especializado en código)
  - `mistral` (Modelo de Mistral AI)
  - `qwen2.5-coder` (Modelo de código de Alibaba)

**Configuración:**
```json
{
  "ai": {
    "provider": "ollama",
    "model": "llama3.2",
    "url": "http://localhost:11434/v1"
  }
}
```

**Variables de Entorno:**
- `OLLAMA_URL` - URL del servidor Ollama personalizado (predeterminado: `http://localhost:11434/v1`)
- `OLLAMA_MODEL` - Modelo a usar

**Notas:**
- No se requiere clave API para instalaciones locales de Ollama
- Soporta URLs personalizadas para servidores Ollama remotos o endpoints compatibles con OpenAI
- El campo `url` puede establecerse a nivel superior (`ai.url`) o específico del proveedor (`ai.ollama.url`)

#### OpenRouter (Múltiples Modelos)

Usa OpenRouter para acceder a múltiples modelos de IA a través de una sola API.

- **Obtener Clave API:** [openrouter.ai/keys](https://openrouter.ai/keys)
- **Modelos Disponibles:** Ver [openrouter.ai/models](https://openrouter.ai/models)

**Configuración:**
```json
{
  "ai": {
    "provider": "openrouter",
    "apiKey": "sk-or-...",
    "model": "anthropic/claude-3-sonnet",
    "url": "https://openrouter.ai/api/v1"
  }
}
```

**Variables de Entorno:**
- `OPENROUTER_API_KEY` - Clave API de OpenRouter
- `OPENROUTER_MODEL` - Modelo a usar
- `OPENROUTER_URL` - URL personalizada de OpenRouter (predeterminado: `https://openrouter.ai/api/v1`)

#### Proveedor Mock (para Pruebas)

El proveedor `mock` es el predeterminado y no requiere clave API. Es útil para probar el flujo de trabajo de IA o para usuarios que no necesitan capacidades de IA reales. Genera consultas simples y predecibles basadas en palabras clave.

**Configuración:**
```json
{
  "ai": {
    "provider": "mock"
  }
}
```

#### Usar Variables de Entorno para Claves API y URLs

Para mejor seguridad, puedes proporcionar claves API y URLs personalizadas mediante variables de entorno en lugar de escribirlas en el archivo de configuración.

**Claves API:**
- **OpenAI:** `OPENAI_API_KEY`
- **Anthropic:** `ANTHROPIC_API_KEY`
- **Google Gemini:** `GEMINI_API_KEY`
- **OpenRouter:** `OPENROUTER_API_KEY`

**URLs Personalizadas:**
- **Ollama:** `OLLAMA_URL` (predeterminado: `http://localhost:11434/v1`)
- **OpenRouter:** `OPENROUTER_URL` (predeterminado: `https://openrouter.ai/api/v1`)

Si se establece una variable de entorno, se utilizará incluso si hay un valor presente en `cqlai.json`.

**Opciones de Configuración:**

| Opción | Tipo | Predeterminado | Descripción |
|--------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Dirección del host de Cassandra |
| `port` | number | `9042` | Puerto de Cassandra |
| `keyspace` | string | `""` | Keyspace predeterminado a usar |
| `username` | string | `""` | Nombre de usuario para autenticación |
| `password` | string | `""` | Contraseña para autenticación |
| `requireConfirmation` | boolean | `true` | Requerir confirmación para comandos peligrosos |
| `consistency` | string | `LOCAL_ONE` | Nivel de consistencia predeterminado (ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE) |
| `pageSize` | number | `100` | Número de filas por página |
| `maxMemoryMB` | number | `10` | Memoria máxima para resultados de consultas en MB |
| `connectTimeout` | number | `10` | Tiempo de espera de conexión en segundos |
| `requestTimeout` | number | `10` | Tiempo de espera de petición en segundos |
| `historyFile` | string | `~/.cqlai/history` | Ruta al archivo de historial de comandos CQL (soporta expansión `~`) |
| `aiHistoryFile` | string | `~/.cqlai/ai_history` | Ruta al archivo de historial de comandos IA (soporta expansión `~`) |
| `debug` | boolean | `false` | Habilitar registro de depuración |

### Ubicaciones de Archivos de Configuración

CQLAI busca archivos de configuración en las siguientes ubicaciones:

**Archivos CQLSHRC:**
1. `$CQLSH_RC` (si se establece la variable de entorno)
2. `~/.cassandra/cqlshrc` (ubicación estándar de cqlsh)
3. `~/.cqlshrc` (ubicación alternativa)

**Archivos JSON de CQLAI:**
1. `./cqlai.json` (directorio de trabajo actual)
2. `~/.cqlai.json` (directorio home del usuario)
3. `~/.config/cqlai/config.json` (directorio de configuración XDG en Linux/macOS)

### Variables de Entorno

Variables de entorno comunes:
- `CQLAI_HOST` o `CASSANDRA_HOST` - Host de Cassandra
- `CQLAI_PORT` o `CASSANDRA_PORT` - Puerto de Cassandra
- `CQLAI_KEYSPACE` - Keyspace predeterminado
- `CQLAI_USERNAME` - Nombre de usuario para autenticación
- `CQLAI_PASSWORD` - Contraseña para autenticación
- `CQLAI_PAGE_SIZE` - Tamaño de paginación en modo batch (predeterminado: 100)
- `CQLSH_RC` - Ruta a archivo CQLSHRC personalizado

### Migración desde cqlsh

Si estás migrando desde `cqlsh`, CQLAI leerá automáticamente tu archivo existente `~/.cassandra/cqlshrc`. No se necesitan cambios para comenzar a usar CQLAI con tu configuración existente de Cassandra.

## Generación de Consultas Potenciada por IA

CQLAI incluye capacidades de IA integradas para convertir lenguaje natural en consultas CQL. Simplemente prefija tu solicitud con `.ai`:

### Ejemplos

```sql
-- Consultas simples
.ai mostrar todos los usuarios
.ai encontrar productos con precio menor a 100
.ai contar pedidos del mes pasado

-- Operaciones complejas
.ai crear una tabla para almacenar comentarios de clientes con id, customer_id, rating y comment
.ai actualizar estado de usuario a inactivo donde last_login sea mayor a 90 días
.ai eliminar todas las sesiones expiradas

-- Exploración de esquema
.ai qué tablas hay en este keyspace
.ai describir la estructura de la tabla users
```

### Cómo Funciona

1. **Entrada en Lenguaje Natural**: Escribe `.ai` seguido de tu solicitud en español
2. **Contexto de Esquema**: CQLAI extrae automáticamente tu esquema actual para proporcionar contexto
3. **Generación de Consulta**: La IA genera un plan de consulta estructurado
4. **Vista Previa y Confirmación**: Revisa el CQL generado antes de la ejecución
5. **Ejecutar o Editar**: Elige ejecutar, editar o cancelar la consulta

### Proveedores de IA Soportados

Configura tu proveedor de IA preferido en `cqlai.json`:

- **OpenAI** (GPT-4, GPT-3.5)
- **Anthropic** (Claude 3)
- **Google Gemini**
- **Ollama** (Modelos locales o APIs compatibles con OpenAI)
- **OpenRouter** (Acceso a múltiples modelos)
- **Mock** (predeterminado, para pruebas sin claves API)

### Características de Seguridad

- **Solo lectura por defecto**: La IA prefiere consultas SELECT a menos que se solicite explícitamente modificar
- **Advertencias de operaciones peligrosas**: Operaciones DROP, DELETE, TRUNCATE muestran advertencias
- **Confirmación requerida**: Operaciones destructivas requieren confirmación adicional
- **Validación de esquema**: Las consultas se validan contra tu esquema actual

## Soporte de Apache Parquet

CQLAI proporciona soporte integral para el formato Apache Parquet, haciéndolo ideal para flujos de trabajo de análisis de datos e integración con ecosistemas de datos modernos.

### Beneficios Clave

- **Almacenamiento Eficiente**: Formato columnar con excelente compresión (50-80% más pequeño que CSV)
- **Análisis Rápido**: Optimizado para consultas analíticas en Spark, Presto y otros motores
- **Preservación de Tipos**: Mantiene tipos de datos de Cassandra incluyendo colecciones y UDTs
- **Listo para Aprendizaje Automático**: Compatibilidad directa con pandas, PyArrow y frameworks de ML
- **Soporte de Streaming**: Streaming eficiente en memoria para conjuntos de datos grandes

### Ejemplos Rápidos

```sql
-- Exportar a Parquet (autodetectado por extensión)
COPY users TO 'users.parquet';

-- Exportar con compresión
COPY events TO 'events.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='ZSTD';

-- Importar desde Parquet
COPY users FROM 'users.parquet';

-- Capturar resultados de consulta en formato Parquet
CAPTURE 'results.parquet' FORMAT='PARQUET';
SELECT * FROM large_table WHERE condition = true;
CAPTURE OFF;
```

### Características Soportadas

- Todos los tipos primitivos de Cassandra (int, text, timestamp, uuid, etc.)
- Tipos de colección (list, set, map)
- Tipos Definidos por el Usuario (UDTs)
- Colecciones congeladas
- Tipos vectoriales para cargas de trabajo de ML (Cassandra 5.0+)
- Múltiples algoritmos de compresión (Snappy, GZIP, ZSTD, LZ4)

Para documentación detallada, consulta [Guía de Soporte de Parquet](docs/PARQUET.md).

## Limitaciones Conocidas

### Salida JSON (CAPTURE JSON y --format json)

Al generar datos como JSON, existen algunas limitaciones debido a cómo el driver gocql subyacente maneja el tipado dinámico:

#### Valores NULL
- **Problema**: Los valores NULL en columnas primitivas (int, boolean, text, etc.) aparecen como valores cero (`0`, `false`, `""`) en lugar de `null`
- **Causa**: El driver gocql devuelve valores cero para NULLs al escanear en tipos dinámicos (`interface{}`)
- **Solución alternativa**: Usa consultas `SELECT JSON` que devuelven JSON apropiado del lado del servidor de Cassandra

#### Tipos Definidos por el Usuario (UDTs)
- **Problema**: Las columnas UDT aparecen como objetos vacíos `{}` en la salida JSON
- **Causa**: El driver gocql no puede deserializar apropiadamente UDTs sin conocimiento en tiempo de compilación de su estructura
- **Solución alternativa**: Usa consultas `SELECT JSON` para serialización apropiada de UDT

#### Ejemplo
```sql
-- SELECT regular (tiene limitaciones)
SELECT * FROM users;
-- Devuelve: {"id": 1, "age": 0, "active": false}  -- age y active podrían ser NULL

-- Usando SELECT JSON (preserva tipos correctamente)
SELECT JSON * FROM users;
-- Devuelve: {"id": 1, "age": null, "active": null}  -- NULLs apropiadamente representados
```

**Nota**: Los tipos complejos (lists, sets, maps, vectors) se preservan apropiadamente en la salida JSON.

## Desarrollo

Para trabajar en `cqlai`, necesitarás Go (≥ 1.24).

#### Configuración

```bash
# Clonar el repositorio
git clone https://github.com/axonops/cqlai.git
cd cqlai

# Instalar dependencias
go mod download
```

#### Compilación

```bash
# Compilar un binario estándar
make build

# Compilar un binario de desarrollo con detección de condiciones de carrera
make build-dev
```

#### Ejecutar Pruebas y Linter

```bash
# Ejecutar todas las pruebas
make test

# Ejecutar pruebas con reporte de cobertura
make test-coverage

# Ejecutar el linter
make lint

# Ejecutar todas las verificaciones (formato, lint, pruebas)
make check
```


## Stack Tecnológico

- **Lenguaje:** Go
- **Framework TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Componentes TUI:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Estilos:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Driver de Cassandra:** [gocql](https://github.com/gocql/gocql)

## Licencia

Este proyecto está licenciado bajo la licencia Apache 2.0. Consulta el archivo LICENSE para más detalles.

Las licencias de dependencias de terceros están disponibles en el directorio [THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/). Para regenerar las atribuciones de licencia, ejecuta `make licenses`.

---

<div align="center">
  <br>
  <p>Desarrollado por</p>
  <img src="./assets/AxonOps-RGB-transparent-small.png" alt="AxonOps" width="200">
</div>
