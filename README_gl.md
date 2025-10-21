# CQLAI - Shell Moderno de Cassandra CQL

<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">
</div>

**CQLAI** é un terminal interactivo rápido e portátil para Cassandra (CQL), construído en Go. Proporciona unha alternativa moderna e fácil de usar a `cqlsh` cunha interface de terminal avanzada, análise de comandos do lado do cliente e funcións de produtividade melloradas.

**As funcións de IA son completamente opcionais** - CQLAI funciona perfectamente como un shell CQL independente sen ningunha configuración de IA ou claves API.

O comando cqlsh orixinal está escrito en Python, o que require que Python estea instalado no sistema. cqlai está compilado nun único binario executable, sen requirir dependencias externas. Este proxecto proporciona binarios para as seguintes plataformas:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


Está construído con [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), e [Lip Gloss](https://github.com/charmbracelet/lipgloss) para a fermosa interface de terminal. Un gran recoñecemento ao equipo do driver gocql de Cassandra por implementar as últimas funcionalidades de Cassandra [gocql](https://github.com/apache/cassandra-gocql-driver)


---

## Estado do Proxecto

**CQLAI está listo para produción** e utilízase activamente en contornas de desenvolvemento, probas e produción con clústeres de Cassandra. A ferramenta proporciona unha alternativa completa e estable a `cqlsh` con características e rendemento mellorados.

### O que Funciona
- Todas as operacións e consultas CQL principais
- Soporte completo de meta-comandos (`DESCRIBE`, `SHOW`, `CONSISTENCY`, etc.)
- Análise de comandos do lado do cliente (lixeiro, sen dependencia de ANTLR)
- Importación/exportación de datos con `COPY TO/FROM` (formatos CSV e Parquet)
- Conexións SSL/TLS e autenticación
- Tipos Definidos polo Usuario (UDTs) e tipos de datos complexos
- Modo batch para scripting e automatización
- Soporte de formato Apache Parquet para intercambio eficiente de datos
- Autocompletado con tabulador para palabras clave CQL, táboas, columnas e keyspaces
- Tamaño de binario pequeno (~43MB, 53% máis pequeno que versións anteriores)
- **Opcional**: Xeración de consultas potenciada por IA (OpenAI, Anthropic, Gemini)

### Próximamente
- Conciencia de contexto de IA mellorada
- Servizo MCP de Cassandra
- Optimizacións de rendemento adicionais

Animámoste a **probar CQLAI hoxe** e axudar a dar forma ao seu desenvolvemento. A túa retroalimentación e contribucións son valiosas para facer deste o mellor shell CQL para a comunidade de Cassandra. Por favor [reporta problemas](https://github.com/axonops/cqlai/issues) ou [contribúe](https://github.com/axonops/cqlai/pulls).

---

## Características

- **Shell CQL Interactivo:** Executa calquera consulta CQL que o teu clúster de Cassandra soporte.
- **Interface de Terminal Enriquecida:**
    - Unha aplicación de terminal de múltiples capas e pantalla completa con búfer de pantalla alternativo (preserva o historial do terminal).
    - Táboa virtualizada e desprazable para resultados con carga automática de datos, prevenindo sobrecarga de memoria en consultas grandes.
    - Modos de navegación avanzados con atallos de teclado estilo vim.
    - Soporte completo de rato incluíndo desprazamento con roda e selección de texto.
    - Barra de estado/pé de páxina fixa mostrando detalles de conexión, latencia de consulta e estado de sesión (consistencia, trazado).
    - Superposicións modais para historial, axuda e autocompletado de comandos.
- **Soporte de Apache Parquet:**
    - Formato de datos columnar de alto rendemento para fluxos de traballo de análise e aprendizaxe automática.
    - Exporta táboas de Cassandra a arquivos Parquet co comando `COPY TO`.
    - Importa arquivos Parquet a Cassandra con inferencia automática de esquema.
    - Conxuntos de datos particionados con estruturas de directorios estilo Hive.
    - Columnas virtuais TimeUUID / timestamp para particionamento intelixente baseado en tempo.
    - Soporte para todos os tipos de datos de Cassandra incluíndo UDTs, coleccións e vectores.
- **Xeración de Consultas Potenciada por IA (Opcional):**
    - Conversión de linguaxe natural a CQL usando provedores de IA (OpenAI, Anthropic, Gemini).
    - Xeración de consultas con conciencia de esquema e contexto automático.
    - Vista previa segura e confirmación antes da execución.
    - Soporte para operacións complexas incluíndo DDL e DML.
    - **Require configuración de clave API** - non necesaria para a funcionalidade principal.
- **Configuración:**
    - Configuración simple mediante `cqlai.json` no directorio actual ou `~/.cqlai.json`.
    - Soporte para conexións SSL/TLS con autenticación por certificado.
- **Binario Único:** Distribuído como un único binario estático sen dependencias externas. Inicio rápido e pegada pequena.

## Instalación

Podes instalar `cqlai` de varias maneiras. Para instrucións detalladas incluíndo xestores de paquetes (APT, YUM) e Docker, consulta a [Guía de Instalación](docs/INSTALLATION.md).

### Binarios Precompilados

Descarga o binario apropiado para o teu sistema operativo e arquitectura desde a páxina de [**Releases**](https://github.com/axonops/cqlai/releases).


### Usando Go

```bash
go install github.com/axonops/cqlai/cmd/cqlai@latest
```

### Desde o Código Fonte

```bash
git clone https://github.com/axonops/cqlai.git
cd cqlai
go build -o cqlai cmd/cqlai/main.go
```

### Usando Docker

```bash
# Construír a imaxe
docker build -t cqlai .

# Executar o contedor
docker run -it --rm --name cqlai-session cqlai --host o-teu-host-cassandra
```

## Uso

### Modo Interactivo

Conectar a un host de Cassandra:
```bash
# Con contrasinal en liña de comandos (non recomendado - visible en ps)
cqlai --host 127.0.0.1 --port 9042 --username cassandra --password cassandra

# Con solicitude de contrasinal (seguro - contrasinal oculto)
cqlai --host 127.0.0.1 --port 9042 -u cassandra
# Password: [entrada oculta]

# Usando variable de contorno (seguro para scripts/contedores)
export CQLAI_PASSWORD=cassandra
cqlai --host 127.0.0.1 -u cassandra
```

Ou usa un arquivo de configuración:
```bash
# Crear configuración desde o exemplo
cp cqlai.json.example cqlai.json
# Editar cqlai.json coa túa configuración, logo executa:
cqlai
```

### Opcións de Liña de Comandos

```bash
cqlai [opcións]
```

#### Opcións de Conexión
| Opción | Curta | Descrición |
|--------|-------|-------------|
| `--host <host>` | | Host de Cassandra (sobrescribe config) |
| `--port <porto>` | | Porto de Cassandra (sobrescribe config) |
| `--keyspace <keyspace>` | `-k` | Keyspace predeterminado (sobrescribe config) |
| `--username <usuario>` | `-u` | Usuario para autenticación |
| `--password <contrasinal>` | `-p` | Contrasinal para autenticación* |
| `--no-confirm` | | Desactivar confirmacións |
| `--connect-timeout <segundos>` | | Tempo de espera de conexión (predeterminado: 10) |
| `--request-timeout <segundos>` | | Tempo de espera de petición (predeterminado: 10) |
| `--debug` | | Habilitar rexistro de depuración |

*\*Nota: O contrasinal pode proporcionarse de tres maneiras:*
1. *Liña de comandos con `-p` (non recomendado - visible na lista de procesos)*
2. *Solicitude interactiva cando se usa `-u` sen `-p` (recomendado)*
3. *Variable de contorno `CQLAI_PASSWORD` (bo para automatización)*

#### Opcións de Modo Batch
| Opción | Curta | Descrición |
|--------|-------|-------------|
| `--execute <declaración>` | `-e` | Executar declaración CQL e saír |
| `--file <arquivo>` | `-f` | Executar CQL desde arquivo e saír |
| `--format <formato>` | | Formato de saída: ascii, json, csv, table |
| `--no-header` | | Non mostrar cabeceiras de columna (CSV) |
| `--field-separator <sep>` | | Separador de campos para CSV (predeterminado: ,) |
| `--page-size <n>` | | Filas por lote (predeterminado: 100) |

#### Opcións Xerais
| Opción | Curta | Descrición |
|--------|-------|-------------|
| `--config-file <ruta>` | | Ruta ao arquivo de configuración (sobrescribe localizacións predeterminadas) |
| `--help` | `-h` | Mostrar mensaxe de axuda |
| `--version` | `-v` | Imprimir versión e saír |

### Exemplos de Modo Batch

Executar declaracións CQL de forma non interactiva (compatible con cqlsh):

```bash
# Executar unha soa declaración
cqlai -e "SELECT * FROM system_schema.keyspaces;"

# Executar desde un arquivo
cqlai -f script.cql

# Entrada por tubería
echo "SELECT * FROM users;" | cqlai

# Controlar formato de saída
cqlai -e "SELECT * FROM users;" --format json
cqlai -e "SELECT * FROM users;" --format csv --no-header

# Controlar tamaño de paxinación
cqlai -e "SELECT * FROM large_table;" --page-size 50
```

### Comandos Básicos

- **Executar CQL:** Escribe calquera declaración CQL e preme Enter.
- **Meta-Comandos:**
  ```sql
  DESCRIBE KEYSPACES;
  USE meu_keyspace;
  DESCRIBE TABLES;
  CONSISTENCY QUORUM;
  TRACING ON;
  PAGING 50;
  EXPAND ON;  -- Modo de saída vertical
  SOURCE 'script.cql';  -- Executar script CQL
  ```
- **Xeración de Consultas Potenciada por IA:**
  ```sql
  .ai Que keyspaces hai?
  .ai Que columnas ten a táboa users?
  .ai crear unha táboa para almacenar inventario de produtos
  .ai eliminar pedidos de máis de 1 ano da táboa orders
  ```

### Atallos de Teclado

#### Navegación e Control
| Atallo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `↑`/`↓` | Navegar historial de comandos | Igual |
| `Ctrl+P`/`Ctrl+N` | Anterior/Seguinte en historial de comandos | Igual |
| `Alt+N` | Mover a seguinte liña en historial | `Option+N` |
| `Tab` | Autocompletar comandos e nomes de táboas/keyspaces | Igual |
| `Ctrl+C` | Limpar entrada / Cancelar paxinación / Cancelar operación (dúas veces para saír) | `⌘+C` ou `Ctrl+C` |
| `Ctrl+D` | Saír da aplicación | `⌘+D` ou `Ctrl+D` |
| `Ctrl+R` | Buscar en historial de comandos | `⌘+R` ou `Ctrl+R` |
| `Esc` | Activar/desactivar modo de navegación / Cancelar paxinación / Pechar modais | Igual |
| `Enter` | Executar comando / Cargar seguinte páxina (durante paxinación) | Igual |

#### Edición de Texto
| Atallo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `Ctrl+A` | Saltar ao inicio da liña | Igual |
| `Ctrl+E` | Saltar ao final da liña | Igual |
| `Ctrl+Esq`/`Ctrl+Der` | Saltar por palabra (ou 20 caracteres) | Igual |
| `PgUp`/`PgDn` (en entrada) | Páxina esq/der en consultas longas | `Fn+↑`/`Fn+↓` |
| `Ctrl+K` | Cortar desde o cursor ata o final da liña | Igual |
| `Ctrl+U` | Cortar desde o inicio ata o cursor | Igual |
| `Ctrl+W` | Cortar palabra cara atrás | Igual |
| `Alt+D` | Eliminar palabra cara adiante | `Option+D` |
| `Ctrl+Y` | Pegar texto cortado previamente | Igual |

#### Cambio de Vista
| Atallo | Acción |
|----------|--------|
| `F2` | Cambiar a vista de consulta/historial |
| `F3` | Cambiar a vista de táboa |
| `F4` | Cambiar a vista de trazas (cando o trazado está habilitado) |
| `F5` | Cambiar a vista de conversa IA |
| `F6` | Activar/desactivar tipos de datos de columna en cabeceiras de táboa |

#### Desprazamento e Navegación de Táboa
| Atallo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `PgUp`/`PgDn` | Desprazar vista por páxina / Cargar máis datos cando estea dispoñible | `Fn+↑`/`Fn+↓` |
| `Espazo` | Cargar seguinte páxina cando haxa máis datos dispoñibles | Igual |
| `Enter` (entrada baleira) | Cargar seguinte páxina cando haxa máis datos dispoñibles | Igual |
| `Alt+↑`/`Alt+↓` | Desprazar vista por unha soa fila (respecta límites de fila) | `Option+↑`/`Option+↓` |
| `Alt+←`/`Alt+→` | Desprazar táboa horizontalmente (táboas anchas) | `Option+←`/`Option+→` |
| `↑`/`↓` | Navegar filas de táboa (cando está en modo navegación) | Igual |

#### Modo de Navegación (Vistas de Táboa/Trazas)
Preme `Esc` para activar/desactivar o modo de navegación cando vexas táboas ou trazas.

| Atallo | Acción en Modo de Navegación |
|----------|---------------------------|
| `j` / `k` | Desprazar abaixo/arriba por unha soa liña |
| `d` / `u` | Desprazar abaixo/arriba por media páxina |
| `g` / `G` | Saltar ao inicio/final de resultados |
| `<` / `>` | Desprazar esq/der por 10 columnas |
| `{` / `}` | Desprazar esq/der por 50 columnas |
| `0` / `$` | Saltar a primeira/última columna |
| `Esc` | Saír do modo de navegación / Cancelar paxinación se está activa |

#### Soporte de Rato
| Acción | Función |
|--------|----------|
| Roda do Rato | Desprazamento vertical con carga automática de datos |
| Alt+Roda do Rato | Desprazamento horizontal en táboas |
| Shift+Roda do Rato | Desprazamento horizontal (alternativa) |
| Ctrl+Roda do Rato | Desprazamento horizontal (alternativa) |
| Shift+Clic+Arrastre | Seleccionar texto para copiar |
| Ctrl+Shift+C | Copiar texto seleccionado ao portapapeis |
| Clic do Medio | Pegar desde o búfer de selección (Linux/Unix) |

**Nota para Usuarios de macOS:**
- A maioría de atallos `Ctrl` funcionan tal cal en macOS, pero tamén podes usar a tecla `⌘` (Comando) como alternativa
- A tecla `Alt` está etiquetada como `Option` nos teclados Mac
- As teclas de función (F1-F6) poden requirir manter premida a tecla `Fn` dependendo da túa configuración de Mac

### Autocompletado con Tabulador

CQLAI proporciona autocompletado intelixente e consciente do contexto para acelerar o teu fluxo de traballo. Preme `Tab` en calquera momento para ver as opcións de autocompletado dispoñibles.

#### Que se Pode Autocompletar

**Palabras Clave e Comandos CQL:**
- Todas as palabras clave CQL: `SELECT`, `INSERT`, `CREATE`, `ALTER`, `DROP`, etc.
- Meta-comandos: `DESCRIBE`, `CONSISTENCY`, `COPY`, `SHOW`, etc.
- Tipos de datos: `TEXT`, `INT`, `UUID`, `TIMESTAMP`, etc.
- Niveis de consistencia: `ONE`, `QUORUM`, `ALL`, `LOCAL_QUORUM`, etc.

**Obxectos de Esquema:**
- Nomes de keyspaces
- Nomes de táboas (dentro do keyspace actual)
- Nomes de columnas (cando o contexto o permite)
- Nomes de tipos definidos polo usuario
- Nomes de funcións e agregados
- Nomes de índices

**Autocompletados Conscientes do Contexto:**
```sql
-- Despois de SELECT, suxire nomes de columnas e palabras clave
SELECT <Tab>           -- Mostra: *, nomes de columnas, DISTINCT, JSON, etc.

-- Despois de FROM, suxire nomes de táboas
SELECT * FROM <Tab>    -- Mostra: táboas dispoñibles no keyspace actual

-- Despois de USE, suxire nomes de keyspaces
USE <Tab>              -- Mostra: keyspaces dispoñibles

-- Despois de DESCRIBE, suxire tipos de obxectos
DESCRIBE <Tab>         -- Mostra: KEYSPACE, TABLE, TYPE, etc.

-- Despois do comando de consistencia
CONSISTENCY <Tab>      -- Mostra: ONE, QUORUM, ALL, etc.
```

**Autocompletado de Rutas de Arquivo:**
```sql
-- Para comandos que aceptan rutas de arquivo
SOURCE '<Tab>          -- Mostra: arquivos no directorio actual
SOURCE '/ruta/<Tab>    -- Mostra: arquivos en /ruta/
```

#### Comportamento do Autocompletado

- **Insensible a Maiúsculas:** Escribe `sel<Tab>` para obter `SELECT`
- **Coincidencia Parcial:** Escribe parte dunha palabra e preme Tab
- **Múltiples Coincidencias:** Cando hai múltiples opcións de autocompletado dispoñibles:
  - Primeiro Tab: Mostra autocompletado en liña se é único
  - Segundo Tab: Mostra todas as opcións dispoñibles nun modal
- **Filtrado Intelixente:** Os autocompletados fíltranse segundo o contexto actual
- **Escape para Cancelar:** Preme `Esc` para pechar o modal de autocompletado

#### Exemplos

```sql
-- Autocompletar nome de táboa
SELECT * FROM us<Tab>
-- Completa a: SELECT * FROM users

-- Autocompletar nivel de consistencia
CONSISTENCY LOC<Tab>
-- Mostra: LOCAL_ONE, LOCAL_QUORUM, LOCAL_SERIAL

-- Autocompletar nomes de columnas despois de SELECT
SELECT id, na<Tab> FROM users
-- Completa a: SELECT id, name FROM users

-- Autocompletar rutas de arquivo para comando SOURCE
SOURCE 'sche<Tab>
-- Completa a: SOURCE 'schema.cql'

-- Autocompletar opcións do comando COPY
COPY users TO 'file.csv' WITH <Tab>
-- Mostra: HEADER, DELIMITER, NULLVAL, PAGESIZE, etc.

-- Mostrar todas as táboas cando existen múltiples
SELECT * FROM <Tab>
-- Mostra modal con: users, orders, products, etc.
```

#### Consellos para Uso Efectivo

1. **Usa Tab liberalmente:** O sistema de autocompletado é intelixente e consciente do contexto
2. **Escribe caracteres mínimos:** A miúdo 2-3 caracteres son suficientes para obter un autocompletado único
3. **Usa para descubrir:** Preme Tab en entrada baleira para ver que está dispoñible
4. **Rutas de arquivo:** Lembra incluír comiñas para autocompletado de rutas de arquivo
5. **Navega autocompletados:** Usa as teclas de frecha para seleccionar entre múltiples opcións

## Comandos Dispoñibles

CQLAI soporta todos os comandos CQL estándar ademais de meta-comandos adicionais para funcionalidade mellorada.

### Comandos CQL
Executa calquera declaración CQL válida soportada polo teu clúster de Cassandra:
- DDL: `CREATE`, `ALTER`, `DROP` (KEYSPACE, TABLE, INDEX, etc.)
- DML: `SELECT`, `INSERT`, `UPDATE`, `DELETE`
- DCL: `GRANT`, `REVOKE`
- Outros: `USE`, `TRUNCATE`, `BEGIN BATCH`, etc.

### Meta-Comandos

Os meta-comandos proporcionan funcionalidade adicional máis alá do CQL estándar:

#### Xestión de Sesión
- **CONSISTENCY** `<nivel>` - Establecer nivel de consistencia (ONE, QUORUM, ALL, etc.)
  ```sql
  CONSISTENCY QUORUM
  CONSISTENCY LOCAL_ONE
  ```

- **PAGING** `<tamaño>` | OFF - Establecer tamaño de paxinación de resultados
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

- **OUTPUT** [FORMATO] - Establecer formato de saída
  ```sql
  OUTPUT          -- Mostrar formato actual
  OUTPUT TABLE    -- Formato de táboa (predeterminado)
  OUTPUT JSON     -- Formato JSON
  OUTPUT EXPAND   -- Formato vertical expandido
  OUTPUT ASCII    -- Formato de táboa ASCII
  ```

#### Descrición de Esquema
- **DESCRIBE** - Mostrar información de esquema
  ```sql
  DESCRIBE KEYSPACES                    -- Listar todos os keyspaces
  DESCRIBE KEYSPACE <nome>              -- Mostrar definición de keyspace
  DESCRIBE TABLES                       -- Listar táboas no keyspace actual
  DESCRIBE TABLE <nome>                 -- Mostrar estrutura de táboa
  DESCRIBE TYPES                        -- Listar tipos definidos polo usuario
  DESCRIBE TYPE <nome>                  -- Mostrar definición de UDT
  DESCRIBE FUNCTIONS                    -- Listar funcións de usuario
  DESCRIBE FUNCTION <nome>              -- Mostrar definición de función
  DESCRIBE AGGREGATES                   -- Listar agregados de usuario
  DESCRIBE AGGREGATE <nome>             -- Mostrar definición de agregado
  DESCRIBE MATERIALIZED VIEWS           -- Listar vistas materializadas
  DESCRIBE MATERIALIZED VIEW <nome>     -- Mostrar definición de vista
  DESCRIBE INDEX <nome>                 -- Mostrar definición de índice
  DESCRIBE CLUSTER                      -- Mostrar información do clúster
  DESC <keyspace>.<táboa>               -- Atallo para descrición de táboa
  ```

#### Exportación/Importación de Datos
- **COPY TO** - Exportar datos de táboa a arquivo CSV ou Parquet
  ```sql
  -- Exportación básica a CSV
  COPY users TO 'users.csv'

  -- Exportar a formato Parquet (autodetectado por extensión)
  COPY users TO 'users.parquet'

  -- Exportar a Parquet con formato e compresión explícitos
  COPY users TO 'data.parquet' WITH FORMAT='PARQUET' AND COMPRESSION='SNAPPY'

  -- Exportar columnas específicas
  COPY users (id, name, email) TO 'users_partial.csv'

  -- Exportar con opcións
  COPY users TO 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|'

  -- Exportar a stdout
  COPY users TO STDOUT WITH HEADER = TRUE

  -- Opcións dispoñibles:
  -- FORMAT = 'CSV'/'PARQUET' -- Formato de saída (predeterminado: CSV, autodetectado)
  -- HEADER = TRUE/FALSE      -- Incluír cabeceiras de columna (só CSV)
  -- DELIMITER = ','          -- Delimitador de campos (só CSV)
  -- NULLVAL = 'NULL'        -- Cadea a usar para valores NULL
  -- PAGESIZE = 1000         -- Filas por páxina para exportacións grandes
  -- COMPRESSION = 'SNAPPY'  -- Para Parquet: SNAPPY, GZIP, ZSTD, LZ4, NONE
  -- CHUNKSIZE = 10000       -- Filas por fragmento para Parquet
  ```

- **COPY FROM** - Importar datos CSV ou Parquet a táboa
  ```sql
  -- Importación básica desde arquivo CSV
  COPY users FROM 'users.csv'

  -- Importar desde arquivo Parquet (autodetectado)
  COPY users FROM 'users.parquet'

  -- Importar desde Parquet con formato explícito
  COPY users FROM 'data.parquet' WITH FORMAT='PARQUET'

  -- Importar con fila de cabeceira (CSV)
  COPY users FROM 'users.csv' WITH HEADER = TRUE

  -- Importar columnas específicas
  COPY users (id, name, email) FROM 'users_partial.csv'

  -- Importar desde stdin
  COPY users FROM STDIN

  -- Importar con opcións personalizadas
  COPY users FROM 'users.csv' WITH HEADER = TRUE AND DELIMITER = '|' AND NULLVAL = 'N/A'

  -- Opcións dispoñibles:
  -- HEADER = TRUE/FALSE      -- Primeira fila contén nomes de columnas
  -- DELIMITER = ','          -- Delimitador de campos
  -- NULLVAL = 'NULL'        -- Cadea representando valores NULL
  -- MAXROWS = -1            -- Máximo de filas a importar (-1 = ilimitado)
  -- SKIPROWS = 0            -- Número de filas iniciais a saltar
  -- MAXPARSEERRORS = -1     -- Máximo de erros de análise permitidos (-1 = ilimitado)
  -- MAXINSERTERRORS = 1000  -- Máximo de erros de inserción permitidos
  -- MAXBATCHSIZE = 20       -- Máximo de filas por inserción batch
  -- MINBATCHSIZE = 2        -- Mínimo de filas por inserción batch
  -- CHUNKSIZE = 5000        -- Filas entre actualizacións de progreso
  -- ENCODING = 'UTF8'       -- Codificación do arquivo
  -- QUOTE = '"'             -- Carácter de comiñas para cadeas
  ```

- **CAPTURE** - Capturar saída de consulta a arquivo (gravación continua)
  ```sql
  CAPTURE 'output.txt'          -- Comezar a capturar a arquivo de texto
  CAPTURE JSON 'output.json'    -- Capturar como JSON
  CAPTURE CSV 'output.csv'      -- Capturar como CSV
  SELECT * FROM users;
  CAPTURE OFF                   -- Deter captura
  ```

- **SAVE** - Gardar resultados de consulta mostrados a arquivo (sen re-executar)
  ```sql
  -- Primeiro executa unha consulta
  SELECT * FROM users WHERE status = 'active';

  -- Logo garda os resultados mostrados en varios formatos:
  SAVE                           -- Diálogo interactivo (elixir formato e nome de arquivo)
  SAVE 'users.csv'               -- Gardar a CSV (formato autodetectado)
  SAVE 'users.json'              -- Gardar a JSON (formato autodetectado)
  SAVE 'users.txt' ASCII         -- Gardar como táboa ASCII
  SAVE 'data.csv' CSV            -- Especificar formato explicitamente

  -- Diferenzas clave con CAPTURE:
  -- - SAVE exporta os resultados mostrados actualmente
  -- - Non necesita re-executar a consulta
  -- - Preserva os datos exactos mostrados no terminal
  -- - Funciona con resultados paxinados (garda só páxinas cargadas)
  ```

#### Visualización de Información
- **SHOW** - Mostrar información de sesión
  ```sql
  SHOW VERSION          -- Mostrar versión de Cassandra
  SHOW HOST            -- Mostrar detalles de conexión actual
  SHOW SESSION         -- Mostrar toda a configuración de sesión
  ```

- **EXPAND** ON | OFF - Activar/desactivar modo de saída expandida
  ```sql
  EXPAND ON            -- Saída vertical (un campo por liña)
  SELECT * FROM users WHERE id = 1;
  EXPAND OFF           -- Saída de táboa normal
  ```

#### Execución de Scripts
- **SOURCE** - Executar scripts CQL desde arquivo
  ```sql
  SOURCE 'schema.cql'           -- Executar script
  SOURCE '/ruta/a/script.cql'   -- Ruta absoluta
  ```

#### Axuda
- **HELP** - Mostrar axuda de comandos
  ```sql
  HELP                 -- Mostrar todos os comandos
  HELP DESCRIBE        -- Axuda para comando específico
  HELP CONSISTENCY     -- Axuda para niveis de consistencia
  ```

### Comandos de IA
- **.ai** `<consulta en linguaxe natural>` - Xerar CQL desde linguaxe natural
  ```sql
  .ai mostrar todos os usuarios con estado activo
  .ai crear unha táboa para almacenar sesións de usuario
  .ai atopar pedidos realizados nos últimos 30 días
  ```

## Configuración

CQLAI soporta múltiples métodos de configuración para máxima flexibilidade e compatibilidade con configuracións existentes de Cassandra.

### Precedencia de Configuración

As fontes de configuración cárganse na seguinte orde (as fontes posteriores sobrescriben as anteriores):

1. **Arquivos CQLSHRC** (para compatibilidade con configuracións cqlsh existentes)
   - `~/.cassandra/cqlshrc` (localización estándar)
   - `~/.cqlshrc` (localización alternativa)
   - `$CQLSH_RC` (se se establece a variable de contorno)

2. **Arquivos de configuración JSON de CQLAI**
   - `./cqlai.json` (directorio actual)
   - `~/.cqlai.json` (directorio home do usuario)
   - `~/.config/cqlai/config.json` (directorio de configuración XDG)

3. **Variables de contorno**
   - `CQLAI_HOST`, `CQLAI_PORT`, `CQLAI_KEYSPACE`, etc.
   - `CASSANDRA_HOST`, `CASSANDRA_PORT` (para compatibilidade)

4. **Bandeiras de liña de comandos** (prioridade máis alta)
   - `--host`, `--port`, `--keyspace`, `--username`, `--password`, etc.

### Compatibilidade con CQLSHRC

CQLAI pode ler arquivos CQLSHRC estándar usados pola ferramenta tradicional `cqlsh`, facendo a migración transparente.

**Seccións CQLSHRC soportadas:**
- `[connection]` - hostname, port, configuración ssl
- `[authentication]` - keyspace, ruta de arquivo de credenciais
- `[auth_provider]` - módulo de autenticación e nome de usuario
- `[ssl]` - configuración de certificados SSL/TLS

**Exemplo de arquivo CQLSHRC:**
```ini
; ~/.cassandra/cqlshrc
[connection]
hostname = cassandra.example.com
port = 9042
ssl = true

[authentication]
keyspace = meu_keyspace
credentials = ~/.cassandra/credentials

[ssl]
certfile = ~/certs/ca.pem
userkey = ~/certs/client-key.pem
usercert = ~/certs/client-cert.pem
validate = true
```

Consulta [CQLSHRC_SUPPORT.md](docs/CQLSHRC_SUPPORT.md) para detalles completos de compatibilidade con CQLSHRC.

### Configuración JSON de CQLAI

Para características avanzadas e configuración de IA, CQLAI usa o seu propio formato JSON:

**Exemplo `cqlai.json`:**
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
    "openai": {
      "apiKey": "sk-...",
      "model": "gpt-4-turbo-preview"
    }
  }
}
```

### Configuración de Provedor de IA

**Nota:** As características de IA son completamente opcionais. CQLAI funciona como un shell CQL completo sen ningunha configuración de IA.

Para habilitar a xeración de consultas potenciada por IA, configura o teu provedor preferido na sección `ai` do teu arquivo `cqlai.json`.

#### OpenAI (GPT-4 e GPT-3.5)

Usa OpenAI para xeración de consultas de alta calidade e propósito xeral. Require unha clave API de OpenAI.

- **Obter Clave API:** [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **Modelos Recomendados:**
  - `gpt-4-turbo-preview` (predeterminado, recomendado para mellores resultados)
  - `gpt-3.5-turbo` (máis rápido, máis económico)

**Configuración:**
```json
{
  "ai": {
    "provider": "openai",
    "openai": {
      "apiKey": "sk-...",
      "model": "gpt-4-turbo-preview"
    }
  }
}
```

#### Anthropic (Claude 3)

Usa Anthropic para modelos potentes e conscientes do contexto. Ideal para consultas complexas e razoamento. Require unha clave API de Anthropic.

- **Obter Clave API:** [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)
- **Modelos Recomendados:**
  - `claude-3-opus-20240229` (máis potente)
  - `claude-3-sonnet-20240229` (predeterminado, rendemento equilibrado)
  - `claude-3-haiku-20240307` (máis rápido)

**Configuración:**
```json
{
  "ai": {
    "provider": "anthropic",
    "anthropic": {
      "apiKey": "sk-ant-...",
      "model": "claude-3-sonnet-20240229"
    }
  }
}
```

#### Google Gemini

Usa Google Gemini para un modelo rápido e capaz de Google. Require unha clave API de Google AI Studio.

- **Obter Clave API:** [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)
- **Modelo Recomendado:**
  - `gemini-pro` (predeterminado)

**Configuración:**
```json
{
  "ai": {
    "provider": "gemini",
    "gemini": {
      "apiKey": "...",
      "model": "gemini-pro"
    }
  }
}
```

#### Provedor Mock (para Probas)

O provedor `mock` é o predeterminado e non require clave API. É útil para probar o fluxo de traballo de IA ou para usuarios que non necesitan capacidades de IA reais. Xera consultas simples e predecibles baseadas en palabras clave.

**Configuración:**
```json
{
  "ai": {
    "provider": "mock"
  }
}
```

#### Usar Variables de Contorno para Claves API

Para mellor seguridade, podes proporcionar claves API mediante variables de contorno en lugar de escribilas no arquivo de configuración.

- **OpenAI:** `OPENAI_API_KEY`
- **Anthropic:** `ANTHROPIC_API_KEY`
- **Google Gemini:** `GEMINI_API_KEY`

Se se establece unha variable de contorno, utilizarase aínda que haxa unha `apiKey` presente en `cqlai.json`.

**Opcións de Configuración:**

| Opción | Tipo | Predeterminado | Descrición |
|--------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Enderezo do host de Cassandra |
| `port` | number | `9042` | Porto de Cassandra |
| `keyspace` | string | `""` | Keyspace predeterminado a usar |
| `username` | string | `""` | Nome de usuario para autenticación |
| `password` | string | `""` | Contrasinal para autenticación |
| `requireConfirmation` | boolean | `true` | Requirir confirmación para comandos perigosos |
| `consistency` | string | `LOCAL_ONE` | Nivel de consistencia predeterminado (ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE) |
| `pageSize` | number | `100` | Número de filas por páxina |
| `maxMemoryMB` | number | `10` | Memoria máxima para resultados de consultas en MB |
| `connectTimeout` | number | `10` | Tempo de espera de conexión en segundos |
| `requestTimeout` | number | `10` | Tempo de espera de petición en segundos |
| `historyFile` | string | `~/.cqlai/history` | Ruta ao arquivo de historial de comandos CQL (soporta expansión `~`) |
| `aiHistoryFile` | string | `~/.cqlai/ai_history` | Ruta ao arquivo de historial de comandos IA (soporta expansión `~`) |
| `debug` | boolean | `false` | Habilitar rexistro de depuración |

### Localizacións de Arquivos de Configuración

CQLAI busca arquivos de configuración nas seguintes localizacións:

**Arquivos CQLSHRC:**
1. `$CQLSH_RC` (se se establece a variable de contorno)
2. `~/.cassandra/cqlshrc` (localización estándar de cqlsh)
3. `~/.cqlshrc` (localización alternativa)

**Arquivos JSON de CQLAI:**
1. `./cqlai.json` (directorio de traballo actual)
2. `~/.cqlai.json` (directorio home do usuario)
3. `~/.config/cqlai/config.json` (directorio de configuración XDG en Linux/macOS)

### Variables de Contorno

Variables de contorno comúns:
- `CQLAI_HOST` ou `CASSANDRA_HOST` - Host de Cassandra
- `CQLAI_PORT` ou `CASSANDRA_PORT` - Porto de Cassandra
- `CQLAI_KEYSPACE` - Keyspace predeterminado
- `CQLAI_USERNAME` - Nome de usuario para autenticación
- `CQLAI_PASSWORD` - Contrasinal para autenticación
- `CQLAI_PAGE_SIZE` - Tamaño de paxinación en modo batch (predeterminado: 100)
- `CQLSH_RC` - Ruta a arquivo CQLSHRC personalizado

### Migración desde cqlsh

Se estás a migrar desde `cqlsh`, CQLAI lerá automaticamente o teu arquivo existente `~/.cassandra/cqlshrc`. Non se necesitan cambios para comezar a usar CQLAI coa túa configuración existente de Cassandra.

## Xeración de Consultas Potenciada por IA

CQLAI inclúe capacidades de IA integradas para converter linguaxe natural en consultas CQL. Simplemente prefixa a túa solicitude con `.ai`:

### Exemplos

```sql
-- Consultas simples
.ai mostrar todos os usuarios
.ai atopar produtos con prezo menor a 100
.ai contar pedidos do mes pasado

-- Operacións complexas
.ai crear unha táboa para almacenar comentarios de clientes con id, customer_id, rating e comment
.ai actualizar estado de usuario a inactivo onde last_login sexa maior a 90 días
.ai eliminar todas as sesións expiradas

-- Exploración de esquema
.ai que táboas hai neste keyspace
.ai describir a estrutura da táboa users
```

### Como Funciona

1. **Entrada en Linguaxe Natural**: Escribe `.ai` seguido da túa solicitude en galego
2. **Contexto de Esquema**: CQLAI extrae automaticamente o teu esquema actual para proporcionar contexto
3. **Xeración de Consulta**: A IA xera un plan de consulta estruturado
4. **Vista Previa e Confirmación**: Revisa o CQL xerado antes da execución
5. **Executar ou Editar**: Elixe executar, editar ou cancelar a consulta

### Provedores de IA Soportados

Configura o teu provedor de IA preferido en `cqlai.json`:

- **OpenAI** (GPT-4, GPT-3.5)
- **Anthropic** (Claude 3)
- **Google Gemini**
- **Mock** (predeterminado, para probas sen claves API)

### Características de Seguridade

- **Só lectura por defecto**: A IA prefire consultas SELECT a menos que se solicite explicitamente modificar
- **Advertencias de operacións perigosas**: Operacións DROP, DELETE, TRUNCATE mostran advertencias
- **Confirmación requirida**: Operacións destrutivas requiren confirmación adicional
- **Validación de esquema**: As consultas valídanse contra o teu esquema actual

## Soporte de Apache Parquet

CQLAI proporciona soporte integral para o formato Apache Parquet, facéndoo ideal para fluxos de traballo de análise de datos e integración con ecosistemas de datos modernos.

### Beneficios Clave

- **Almacenamento Eficiente**: Formato columnar con excelente compresión (50-80% máis pequeno que CSV)
- **Análise Rápida**: Optimizado para consultas analíticas en Spark, Presto e outros motores
- **Preservación de Tipos**: Mantén tipos de datos de Cassandra incluíndo coleccións e UDTs
- **Listo para Aprendizaxe Automática**: Compatibilidade directa con pandas, PyArrow e frameworks de ML
- **Soporte de Streaming**: Streaming eficiente en memoria para conxuntos de datos grandes

### Exemplos Rápidos

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

- Todos os tipos primitivos de Cassandra (int, text, timestamp, uuid, etc.)
- Tipos de colección (list, set, map)
- Tipos Definidos polo Usuario (UDTs)
- Coleccións conxeladas
- Tipos vectoriais para cargas de traballo de ML (Cassandra 5.0+)
- Múltiples algoritmos de compresión (Snappy, GZIP, ZSTD, LZ4)

Para documentación detallada, consulta [Guía de Soporte de Parquet](docs/PARQUET.md).

## Limitacións Coñecidas

### Saída JSON (CAPTURE JSON e --format json)

Ao xerar datos como JSON, existen algunhas limitacións debido a como o driver gocql subxacente manexa o tipado dinámico:

#### Valores NULL
- **Problema**: Os valores NULL en columnas primitivas (int, boolean, text, etc.) aparecen como valores cero (`0`, `false`, `""`) en lugar de `null`
- **Causa**: O driver gocql devolve valores cero para NULLs ao escanear en tipos dinámicos (`interface{}`)
- **Solución alternativa**: Usa consultas `SELECT JSON` que devolven JSON apropiado do lado do servidor de Cassandra

#### Tipos Definidos polo Usuario (UDTs)
- **Problema**: As columnas UDT aparecen como obxectos baleiros `{}` na saída JSON
- **Causa**: O driver gocql non pode deserializar apropiadamente UDTs sen coñecemento en tempo de compilación da súa estrutura
- **Solución alternativa**: Usa consultas `SELECT JSON` para serialización apropiada de UDT

#### Exemplo
```sql
-- SELECT regular (ten limitacións)
SELECT * FROM users;
-- Devolve: {"id": 1, "age": 0, "active": false}  -- age e active poderían ser NULL

-- Usando SELECT JSON (preserva tipos correctamente)
SELECT JSON * FROM users;
-- Devolve: {"id": 1, "age": null, "active": null}  -- NULLs apropiadamente representados
```

**Nota**: Os tipos complexos (lists, sets, maps, vectors) presérvanse apropiadamente na saída JSON.

## Desenvolvemento

Para traballar en `cqlai`, necesitarás Go (≥ 1.24).

#### Configuración

```bash
# Clonar o repositorio
git clone https://github.com/axonops/cqlai.git
cd cqlai

# Instalar dependencias
go mod download
```

#### Compilación

```bash
# Compilar un binario estándar
make build

# Compilar un binario de desenvolvemento con detección de condicións de carreira
make build-dev
```

#### Executar Probas e Linter

```bash
# Executar todas as probas
make test

# Executar probas con reporte de cobertura
make test-coverage

# Executar o linter
make lint

# Executar todas as verificacións (formato, lint, probas)
make check
```


## Stack Tecnolóxico

- **Linguaxe:** Go
- **Framework TUI:** [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Compoñentes TUI:** [Bubbles](https://github.com/charmbracelet/bubbles)
- **Estilos:** [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- **Driver de Cassandra:** [gocql](https://github.com/gocql/gocql)

## Licenza

Este proxecto está licenciado baixo a licenza Apache 2.0. Consulta o arquivo LICENSE para máis detalles.

As licenzas de dependencias de terceiros están dispoñibles no directorio [THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/). Para rexenerar as atribucións de licenza, executa `make licenses`.

---

<div align="center">
  <br>
  <p>Desenvolvido por</p>
  <img src="./assets/AxonOps-RGB-transparent-small.png" alt="AxonOps" width="200">
</div>
