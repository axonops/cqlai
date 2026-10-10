<div align="center">
  <img src="./assets/cqlai-logo.svg" alt="CQLAI Logo" width="400">

  # CQLAI - Shell Moderno de Cassandra® CQL

  [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
  [![Go Version](https://img.shields.io/github/go-mod/go-version/axonops/cqlai)](https://golang.org/)
  [![GitHub Issues](https://img.shields.io/github/issues/axonops/cqlai)](https://github.com/axonops/cqlai/issues)
  [![GitHub Discussions](https://img.shields.io/github/discussions/axonops/cqlai)](https://github.com/axonops/cqlai/discussions)
  [![GitHub Stars](https://img.shields.io/github/stars/axonops/cqlai)](https://github.com/axonops/cqlai/stargazers)
</div>

**CQLAI** é un terminal interactivo rápido e portátil para Cassandra (CQL), construído en Go. Proporciona unha alternativa moderna e fácil de usar a `cqlsh` cunha interface de terminal avanzada, análise de comandos do lado do cliente e funcións de produtividade melloradas.

**As funcións de IA son completamente opcionais** - CQLAI funciona perfectamente como un shell CQL independente sen ningunha configuración de IA ou claves API.

<div align="center">
  <video src="https://github.com/user-attachments/assets/334bd302-3152-4f48-9d2d-ed617e8d86d3" controls width="100%" style="max-width: 800px;">
    Your browser does not support the video tag.
  </video>
</div>

<div align="center">

### 100% Gratuíto e de Código Aberto
**Sen custos ocultos • Sen niveis premium • Sen claves de licenza**

Desenvolvemento impulsado pola comunidade con total transparencia

</div>

O comando cqlsh orixinal no proxecto [Apache Cassandra](https://cassandra.apache.org/) está escrito en Python, o que require que Python estea instalado no sistema. cqlai está compilado nun único binario executable, sen requirir dependencias externas. Este proxecto proporciona binarios para as seguintes plataformas:

- Linux x86-64
- macOS x86-64
- Windows x86-64
- Linux aarch64
- macOS arm64


Está construído con [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Bubbles](https://github.com/charmbracelet/bubbles), e [Lip Gloss](https://github.com/charmbracelet/lipgloss) para a fermosa interface de terminal. Un gran recoñecemento ao equipo do driver gocql de Cassandra por implementar as últimas funcionalidades de Cassandra [gocql](https://github.com/apache/cassandra-gocql-driver)

---

## Táboa de Contidos

- [Estado do Proxecto](#estado-do-proxecto)
- [Características](#características)
- [Instalación](#instalación)
- [Uso](#uso)
  - [Modo Interactivo](#modo-interactivo)
  - [Opcións de Liña de Comandos](#opcións-de-liña-de-comandos)
  - [Exemplos de Modo Batch](#exemplos-de-modo-batch)
  - [Comandos Básicos](#comandos-básicos)
  - [Atallos de Teclado](#atallos-de-teclado)
  - [Autocompletado con Tabulador](#autocompletado-con-tabulador)
- [Comandos Dispoñibles](#comandos-dispoñibles)
- [Configuración](#configuración)
  - [Precedencia de Configuración](#precedencia-de-configuración)
  - [Compatibilidade con CQLSHRC](#compatibilidade-con-cqlshrc)
  - [Configuración JSON de CQLAI](#configuración-json-de-cqlai)
  - [Configuración de Provedor de IA](#configuración-de-provedor-de-ia)
    - [OpenAI](#openai-gpt-4-e-gpt-35)
    - [Anthropic](#anthropic-claude-3)
    - [Google Gemini](#google-gemini)
    - [Synthetic](#synthetic-múltiples-modelos-de-código-aberto)
    - [Ollama](#ollama-modelos-locais)
    - [OpenRouter](#openrouter-múltiples-modelos)
    - [Provedor Mock](#provedor-mock-para-probas)
- [Xeración de Consultas Potenciada por IA](#xeración-de-consultas-potenciada-por-ia)
- [Servidor MCP](#servidor-mcp)
- [Soporte de Apache Parquet](#soporte-de-apache-parquet)
- [Limitacións Coñecidas](#limitacións-coñecidas)
- [Desenvolvemento](#desenvolvemento)
- [Stack Tecnolóxico](#stack-tecnolóxico)
- [Agradecementos](#agradecementos)
- [Comunidade e Soporte](#comunidade-e-soporte)
- [Licenza](#licenza)
- [Avisos Legais](#avisos-legais)

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
- Autocompletado que constrúe a sentenza enteira - propiedades de táboa, todos os `CREATE` e DML - e indica que escribir onde un nome non se pode consultar
- Navegador de esquema (`F3`): os keyspaces e as táboas do clúster coas súas definicións, que detecta un cambio feito noutra ventá
- Conexións gardadas: CQLAI arranca sen clúster e conéctase a unha con nome desde o menú `FILE`
- Unha ventá `PREFERENCES` que edita `cqlai.json` desde o propio shell
- Rato en todas partes: lapelas e axustes clicables, arrastrar para seleccionar e copiar, e un menú `FILE`
- **Opcional**: IA que escribe CQL en linguaxe natural, le unha traza de consulta e revisa a definición dunha táboa ([OpenAI](https://openai.com/), [Anthropic](https://www.anthropic.com/), [Google Gemini](https://ai.google.dev/), [Ollama](https://ollama.ai/), [OpenRouter](https://openrouter.ai/))

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
    - Conversión de linguaxe natural a CQL usando provedores de IA ([OpenAI](https://openai.com/), [Anthropic](https://www.anthropic.com/), [Google Gemini](https://ai.google.dev/), [Synthetic](https://synthetic.new/)).
    - Xeración de consultas con conciencia de esquema e contexto automático.
    - Vista previa segura e confirmación antes da execución.
    - Soporte para operacións complexas incluíndo DDL e DML.
    - **Require configuración de clave API** - non necesaria para a funcionalidade principal.
- **Configuración:**
    - Configuración simple mediante `cqlai.json` no directorio actual ou `~/.cassandra/cqlai.json`, xunto a `cqlshrc`.
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
| `--no-confirm` | | Desactivar confirmacións para comandos destrutivos (DROP, DELETE, TRUNCATE) |
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
| `Ctrl+Q` | Saír, preguntando antes - igual que `FILE > QUIT` | `Ctrl+Q` (non `⌘+Q`, que pecha o propio terminal) |
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

As lapelas lense de esquerda a dereita, e as teclas seguen a mesma orde.

| Atallo | Acción |
|----------|--------|
| `F2` | Consola: o que escribiches e o que respondeu |
| `F3` | Esquema: os keyspaces e as táboas, coas súas definicións |
| `F4` | Resultados: a última consulta, no formato de `OUTPUT` |
| `F5` | Trazas: a outra lapela de `RESULTS`, cando o trazado está habilitado |
| `F6` | Chat: a conversa coa IA |

#### O menú FILE

`FILE` (Alt+F), na liña de lapelas, reúne o que traballa con ficheiros:

```
╭──────────────╮
│ SAVE RESULTS │
│ SOURCE       │
│ COPY TO      │
│ COPY FROM    │
│──────────────│
│ AUTOSAVE     │
│ PREFERENCES  │
│──────────────│
│ QUIT         │
╰──────────────╯
```

`SAVE RESULTS` escribe o que hai na pantalla. `AUTOSAVE` garda cada consulta a
partir dese momento. `SOURCE` executa o CQL dun ficheiro, e `COPY TO`/`COPY
FROM` moven unha táboa enteira. Baixo a primeira liña, `PREFERENCES` edita os
axustes cos que arranca CQLAI; baixo a segunda, `QUIT` sae, preguntando antes.

As ventás que piden unha ruta abren un explorador de ficheiros no teu directorio
persoal. Tab completa, `..` sobe un nivel, e a roda despraza a lista.

#### O explorador de esquema

`SCHEMA` (F3) amosa os keyspaces e as táboas do clúster como unha árbore á
esquerda, e a definición do seleccionado á dereita. Ao premer nun keyspace
amósase a súa definición e desprégase; ao premer de novo prégase. Ao premer
nunha táboa amósase o seu `CREATE TABLE`.

As frechas percorren a árbore (dereita desprega, esquerda prega) e a roda
despraza o panel que estea baixo o punteiro. O resto segue indo ao prompt, así
que podes escribir unha consulta mentres miras a táboa da que trata.

A cabeceira sobre a árbore é un filtro. Preme `Arriba` dende a primeira fila da
árbore, `/` co prompt baleiro, ou fai clic na cabeceira, e escribe: a árbore amosa só os keyspaces e as táboas cuxo
nome contén o escrito, sen distinguir maiúsculas. Unha táboa amósase baixo o
seu keyspace, sen ter que despregalo antes. Selecciónase a primeira
coincidencia. `Abaixo` ou `Enter` volve á árbore, devolve o teclado ao prompt e
mantén o filtro; `Esc` bórrao e volve amosar a
árbore enteira, co atopado aínda seleccionado.

#### Desprazamento e Navegación de Táboa
| Atallo | Acción | Alternativa macOS |
|----------|--------|-------------------|
| `PgUp`/`PgDn` | Desprazar vista por páxina / Cargar máis datos cando estea dispoñible | `Fn+↑`/`Fn+↓` |
| `Espazo` | Cargar seguinte páxina cando haxa máis datos dispoñibles | Igual |
| `Enter` (entrada baleira) | Cargar seguinte páxina cando haxa máis datos dispoñibles | Igual |
| `Alt+↑`/`Alt+↓` | Desprazar vista por unha soa fila (respecta límites de fila) | `Option+↑`/`Option+↓` |
| `Alt+←`/`Alt+→` | Desprazar táboa horizontalmente (táboas anchas) | `Option+←`/`Option+→` |
| `↑`/`↓` | Desprazan os resultados nas vistas de táboa e traza. Na vista normal percorren o historial de comandos; `Ctrl+P` recupera o historial desde calquera vista | Igual |

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
cqlai deixa os botóns do rato ao teu terminal, así que seleccionar texto, pegar
co botón dereito e pegar co botón central funcionan igual ca en calquera outro
programa. Non fai falta ningunha tecla modificadora.

A roda chega a través do modo de desprazamento alternativo, no que o terminal
converte os xiros da roda en pulsacións de `↑`/`↓`. É dicir, a roda despraza o
mesmo que `↑` e `↓`: os resultados na vista de táboa, a traza na vista de traza
e a conversa na vista de IA.

| Acción | Función |
|--------|----------|
| Roda do Rato | Desprazamento vertical con carga automática de datos |
| Clic+Arrastre | Seleccionar texto (a selección propia do teu terminal) |
| Clic Dereito | Pegar (se o teu terminal o ten asignado así) |
| Clic do Medio | Pegar desde o búfer de selección (Linux/Unix) |

Para desprazar táboas anchas en horizontal usa `Alt+←`/`Alt+→`, ou `<` e `>` en
modo navegación. O desprazamento horizontal coa roda non está dispoñible, porque
o terminal só informa dos xiros verticais como pulsacións de tecla.

**Nota para Usuarios de macOS:**
- A maioría de atallos `Ctrl` funcionan tal cal en macOS, pero tamén podes usar a tecla `⌘` (Comando) como alternativa
- A tecla `Alt` está etiquetada como `Option` nos teclados Mac
- Por defecto macOS usa `Option` para escribir caracteres especiais, así que `Option+F`
  escribe `ƒ` en lugar de enviar `Alt+F`. CQLAI le `ƒ`, `˙`, `∂` e `∫` como `Alt+F`,
  `Alt+H`, `Alt+D` e `Alt+B`, así que eses catro funcionan sen ningún axuste. `å` non se
  toca, porque é unha letra en dinamarqués, noruegués e sueco, así que `Option+A` non abre a IA.
- Para todos os atallos con `Alt`, faga que `Option` envíe `Alt` (Meta) no seu terminal:
  - iTerm2: Settings > Profiles > Keys > Left Option key: `Esc+`
  - Terminal.app: Settings > Profiles > Keyboard > Use Option as Meta key
  - Ghostty: `macos-option-as-alt = true`
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

**Formas das Sentenzas:**

CREATE TABLE, INDEX, MATERIALIZED VIEW, TYPE, FUNCTION, AGGREGATE, TRIGGER,
KEYSPACE, ROLE e USER complétanse palabra a palabra, a traves das suas palabras
clave, das suas parenteses e das suas opcions: `CREATE FUNCTION` ofrece dende a
lista de argumentos ata `CALLED ON NULL INPUT` e a linguaxe.

Tamen SELECT, INSERT, UPDATE, DELETE e BATCH: tras unha columna nun WHERE
aparecen os operadores (`=`, `IN`, `CONTAINS KEY`, `BETWEEN`, `LIKE`,
`IS NOT NULL`), tras a taboa as clausulas que poden seguila na sua orde, e nada
cando a sentenza xa esta completa.

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

-- Despois de WITH, suxire as opcions de taboa e logo os seus valores
... WITH <Tab>                        -- Mostra: compaction = , gc_grace_seconds = , etc.
... WITH compaction = {<Tab>          -- Mostra: as claves que toma o mapa
... WITH compaction = {'class': <Tab> -- Mostra: as estratexias de compactacion
... WITH gc_grace_seconds = <Tab>     -- Mostra: <seconds>
... WITH CLUSTERING ORDER BY (b <Tab> -- Mostra: ASC, DESC

-- Un indice completase ata o seu obxectivo, a sua implementacion e as suas opcions
CREATE INDEX i ON t (<Tab>           -- Mostra: keys(, values(, entries(, full(, <column name>
... (email) USING <Tab>              -- Mostra: 'sai', 'StorageAttachedIndex', etc.
... USING 'sai' WITH OPTIONS = {<Tab> -- Mostra: as opcions dese indice
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
- **Que Escribir:** Onde a seguinte palabra é túa para inventala - o nome dun keyspace, unha táboa, unha columna, ou un valor - a lista dío en vez de quedar baleira: `<table name>`, `<column name>`, `<value>`. Esa nota móstrase en cursiva e nunca se insire no prompt.

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
5. **Navega autocompletados:** Usa as teclas de frecha para seleccionar entre múltiples opcións, ou fai clic nunha para usala. A roda do rato percorre a lista

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
  -- MAXREQUESTS = 6         -- Traballadores de lotes concorrentes (paralelismo)
  -- MINBATCHSIZE = 2        -- Mínimo de filas por inserción batch
  -- CHUNKSIZE = 5000        -- Filas entre actualizacións de progreso
  -- ENCODING = 'UTF8'       -- Codificación do arquivo
  -- QUOTE = '"'             -- Carácter de comiñas para cadeas
  ```

- **AUTOSAVE** - Gardar a saída de cada consulta nun directorio, segundo se executa
  ```sql
  AUTOSAVE '/exports/'           -- Cada consulta como arquivo de texto
  AUTOSAVE JSON '/exports/'      -- Un arquivo JSON por consulta
  AUTOSAVE CSV '/exports/'       -- Un arquivo CSV por consulta
  SELECT * FROM users;
  AUTOSAVE OFF                   -- Deter
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

  -- Diferenzas clave con AUTOSAVE:
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
   - `~/.cassandra/cqlai.json` (xunto a `cqlshrc`, e onde se crea un novo)
   - `~/.cqlai.json` (directorio home do usuario, onde ía antes)
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
  "maxMemoryMB": 10,
  "connectTimeout": 10,
  "requestTimeout": 10,
  "debug": false,
  "historyFile": "~/.cassandra/cqlai_history",
  "aiHistoryFile": "~/.cassandra/cqlai_ai_history",
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

**Nota:** Tamén podes usar o campo `url` para sobrescribir o endpoint da API para APIs compatibles con OpenAI:
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "a-túa-clave-api",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
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
    "apiKey": "sk-...",
    "model": "gpt-4-turbo-preview"
  }
}
```

#### Anthropic (Claude 3)

Usa Anthropic para modelos potentes e conscientes do contexto. Ideal para consultas complexas e razoamento. Require unha clave API de Anthropic.

- **Obter Clave API:** [console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)
- **Modelos Recomendados:**
  - `claude-opus-5` (predeterminado, o máis capaz)
  - `claude-sonnet-5` (máis rápido, menos custoso)
  - `claude-haiku-4-5` (o máis rápido)

**Configuración:**
```json
{
  "ai": {
    "provider": "anthropic",
    "apiKey": "sk-ant-...",
    "model": "claude-opus-5"
  }
}
```

#### Google Gemini

Usa Google Gemini para un modelo rápido e capaz de Google. Require unha clave API de Google AI Studio.

- **Obter Clave API:** [aistudio.google.com/app/apikey](https://aistudio.google.com/app/apikey)
- **Modelo Recomendado:**
  - `gemini-3.8-flash` (predeterminado)

**Configuración:**
```json
{
  "ai": {
    "provider": "gemini",
    "apiKey": "...",
    "model": "gemini-3.8-flash"
  }
}
```

#### Synthetic (Múltiples Modelos de Código Aberto)

Usa Synthetic para acceder a unha ampla selección de modelos de IA de código aberto a prezos moi razoables. Synthetic proporciona unha API compatible con OpenAI que facilita traballar con varios modelos de código aberto.

- **Comezar:** [synthetic.new](https://synthetic.new/)
- **Documentación de API:** [dev.synthetic.new/docs](https://dev.synthetic.new/docs)
- **Modelo Recomendado:**
  - `hf:Qwen/Qwen3-235B-A22B-Instruct-2507` (recomendado, aínda que non probamos exhaustivamente todos os modelos)
- **Modelos Dispoñibles:** Ver [Always-On Models](https://dev.synthetic.new/docs/api/models#always-on-models)

**Configuración:**
```json
{
  "ai": {
    "provider": "openai",
    "apiKey": "a-túa-clave-api-synthetic",
    "url": "https://api.synthetic.new/openai/v1",
    "model": "hf:Qwen/Qwen3-235B-A22B-Instruct-2507"
  }
}
```

**Beneficios Clave:**
- Acceso a unha ampla variedade de modelos de código aberto
- Prezos rendibles
- API compatible con OpenAI para fácil integración
- Sen dependencia de provedor

**Notas:**
- Synthetic presenta unha interface compatible con OpenAI, polo que usas o provedor `openai` na túa configuración
- O campo `url` sobrescribe o endpoint de OpenAI predeterminado para apuntar a Synthetic
- Requírese unha clave API - obtena de [synthetic.new](https://synthetic.new/)

#### Ollama (Modelos Locais)

Usa Ollama para executar modelos de IA localmente ou conectarte a APIs compatibles con OpenAI. Ollama permíteche executar modelos de linguaxe potentes no teu propio hardware sen enviar datos a servizos externos.

- **Comezar:** [ollama.ai](https://ollama.ai)
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

**Variables de Contorno:**
- `OLLAMA_URL` - URL do servidor Ollama personalizado (predeterminado: `http://localhost:11434/v1`)
- `OLLAMA_MODEL` - Modelo a usar

**Notas:**
- Non se require clave API para instalacións locais de Ollama
- Soporta URLs personalizadas para servidores Ollama remotos ou endpoints compatibles con OpenAI
- O campo `url` pode establecerse a nivel superior (`ai.url`) ou específico do provedor (`ai.ollama.url`)

#### OpenRouter (Múltiples Modelos)

Usa OpenRouter para acceder a múltiples modelos de IA a través dunha soa API.

- **Obter Clave API:** [openrouter.ai/keys](https://openrouter.ai/keys)
- **Modelos Dispoñibles:** Ver [openrouter.ai/models](https://openrouter.ai/models)

**Configuración:**
```json
{
  "ai": {
    "provider": "openrouter",
    "apiKey": "sk-or-...",
    "model": "anthropic/claude-opus-5",
    "url": "https://openrouter.ai/api/v1"
  }
}
```

**Variables de Contorno:**
- `OPENROUTER_API_KEY` - Clave API de OpenRouter
- `OPENROUTER_MODEL` - Modelo a usar
- `OPENROUTER_URL` - URL personalizada de OpenRouter (predeterminado: `https://openrouter.ai/api/v1`)

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

#### Usar Variables de Contorno para Claves API e URLs

Para mellor seguridade, podes proporcionar claves API e URLs personalizadas mediante variables de contorno en lugar de escribilas no arquivo de configuración.

**Claves API:**
- **OpenAI:** `OPENAI_API_KEY`
- **Anthropic:** `ANTHROPIC_API_KEY`
- **Google Gemini:** `GEMINI_API_KEY`
- **OpenRouter:** `OPENROUTER_API_KEY`

**URLs Personalizadas:**
- **Ollama:** `OLLAMA_URL` (predeterminado: `http://localhost:11434/v1`)
- **OpenRouter:** `OPENROUTER_URL` (predeterminado: `https://openrouter.ai/api/v1`)

Se se establece unha variable de contorno, utilizarase aínda que haxa un valor presente en `cqlai.json`.

**Opcións de Configuración:**

| Opción | Tipo | Predeterminado | Descrición |
|--------|------|---------|-------------|
| `host` | string | `127.0.0.1` | Enderezo do host de Cassandra |
| `port` | number | `9042` | Porto de Cassandra |
| `keyspace` | string | `""` | Keyspace predeterminado a usar |
| `username` | string | `""` | Nome de usuario para autenticación |
| `password` | string | `""` | Contrasinal para autenticación |
| `requireConfirmation` | boolean | `true` | Requirir confirmación para comandos destrutivos (DROP, DELETE, TRUNCATE) |
| `consistency` | string | `LOCAL_ONE` | Nivel de consistencia predeterminado (ANY, ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE) |
| `pageSize` | number | `100` | Número de filas por páxina |
| `maxMemoryMB` | number | `10` | Memoria máxima para resultados de consultas en MB |
| `connectTimeout` | number | `10` | Tempo de espera de conexión en segundos |
| `requestTimeout` | number | `10` | Tempo de espera de petición en segundos |
| `historyFile` | string | `~/.cassandra/cqlai_history` | Ruta ao arquivo de historial de comandos CQL (soporta expansión `~`). Móvese desde `~/.cqlai/history` a primeira vez que se usa |
| `aiHistoryFile` | string | `~/.cassandra/cqlai_ai_history` | Ruta ao arquivo de historial de comandos IA (soporta expansión `~`). Móvese desde `~/.cqlai/ai_history` a primeira vez que se usa |
| `debug` | boolean | `false` | Habilitar rexistro de depuración |

### Localizacións de Arquivos de Configuración

CQLAI busca arquivos de configuración nas seguintes localizacións:

**Arquivos CQLSHRC:**
1. `$CQLSH_RC` (se se establece a variable de contorno)
2. `~/.cassandra/cqlshrc` (localización estándar de cqlsh)
3. `~/.cqlshrc` (localización alternativa)

**Arquivos JSON de CQLAI:**
1. `./cqlai.json` (directorio de traballo actual)
2. `~/.cassandra/cqlai.json` (xunto a `cqlshrc`)
3. `~/.cqlai.json` (directorio home do usuario)
4. `~/.config/cqlai/config.json` (directorio de configuración XDG en Linux/macOS)

### Variables de Contorno

Variables de contorno comúns:
- `CQLAI_HOST` ou `CASSANDRA_HOST` - Host de Cassandra
- `CQLAI_PORT` ou `CASSANDRA_PORT` - Porto de Cassandra
- `CQLAI_KEYSPACE` - Keyspace predeterminado
- `CQLAI_USERNAME` - Nome de usuario para autenticación
- `CQLAI_PASSWORD` - Contrasinal para autenticación
- `CQLAI_PAGE_SIZE` - Tamaño de paxinación en modo batch (predeterminado: 100)
- `CQLAI_NO_CONFIRM` - Establecer a `true` ou `1` para desactivar confirmacións de comandos destrutivos
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

- **[OpenAI](https://openai.com/)** (GPT-4, GPT-3.5)
- **[Anthropic](https://www.anthropic.com/)** (Claude 3)
- **[Google Gemini](https://ai.google.dev/)**
- **[Synthetic](https://synthetic.new/)** (Múltiples modelos de código aberto)
- **[Ollama](https://ollama.ai/)** (Modelos locais ou APIs compatibles con OpenAI)
- **[OpenRouter](https://openrouter.ai/)** (Acceso a múltiples modelos)
- **Mock** (predeterminado, para probas sen claves API)

### Características de Seguridade

- **Só lectura por defecto**: A IA prefire consultas SELECT a menos que se solicite explicitamente modificar
- **Advertencias de operacións perigosas**: Operacións DROP, DELETE, TRUNCATE mostran advertencias
- **Confirmación requirida**: Operacións destrutivas requiren confirmación adicional
- **Validación de esquema**: As consultas valídanse contra o teu esquema actual

### Desactivar Confirmacións

Para automatización e scripts, podes desactivar as confirmacións para comandos destrutivos (DROP, DELETE, TRUNCATE) usando calquera destes métodos:

1. **Bandeira de liña de comandos**:
   ```bash
   cqlai --no-confirm -e "TRUNCATE my_table;"
   ```

2. **Variable de contorno**:
   ```bash
   export CQLAI_NO_CONFIRM=true
   cqlai -e "DROP TABLE old_data;"
   ```

3. **Arquivo de configuración** (`cqlai.json`):
   ```json
   {
     "requireConfirmation": false
   }
   ```

**Nota**: Usar con precaución en contornos de produción. Estas configuracións desactivan as confirmacións de seguridade que axudan a previr perda accidental de datos.

## Servidor MCP

`cqlai mcp` permite que un asistente de IA traballe cun clúster de Cassandra a
través de CQLAI. O asistente conéctase mediante o
[Model Context Protocol](https://modelcontextprotocol.io) (MCP). CQLAI mantén a
conexión, executa as sentenzas e devolve os resultados en JSON.

O asistente é o cliente MCP. Neste modo CQLAI non chama a ningún provedor de IA,
e non fai falta configurar ningún.

O servidor nunca cambia datos nin esquema. Cando fai falta un cambio, o modelo
propono: CQLAI di que fará, e ti decides se o executas.

### Arrincalo

Hai dúas formas de executalo.

**`cqlai mcp`** abre a shell como sempre e serve MCP desde ela, en
`http://127.0.0.1:7845/mcp` agás que se configure outro enderezo. Serve a conexión que escollas en `FILE > CONNECT`,
cos axustes desa conexión, e segue a túa elección cando escolles outra. Arrinca
haxa ou non un clúster dispoñible; se non o hai, abre `CONNECT` para que
escollas un. A barra de estado amosa `MCP: :7845` mentres serve.

`FILE > MCP SERVER` imprime o que hai que engadir á configuración do cliente,
listo para copiar:

```json
{
  "mcpServers": {
    "cqlai": {
      "type": "http",
      "url": "http://127.0.0.1:7845/mcp"
    }
  }
}
```

**O token.** Está desactivado agás que o actives: marca `Require token` en
MCP SERVER de PREFERENCES, pon `"token": true` no bloque `mcp` de
`cqlai.json`, ou arrinca con `cqlai mcp --token`. Entón cada petición ten que
levalo, e a configuración do cliente de arriba leva unha entrada `headers`
con el:

```json
"headers": { "Authorization": "Bearer <token>" }
```

O token está en `~/.cassandra/cqlai_mcp_token`, que só ti podes
ler. O ficheiro créase a primeira vez que `cqlai mcp` arrinca co token activado, e o token é o
mesmo cada vez que arrinca a shell, así que a configuración do cliente non
cambia. Hai dúas formas de obtelo:

- Na shell: `FILE > MCP SERVER` imprime a configuración do cliente co token xa
  posto na cabeceira `Authorization`.
- Nun terminal: `cat ~/.cassandra/cqlai_mcp_token`

CQLAI 0.3.3 e anteriores gardaban o token en `~/.cqlai_mcp_token`, e o
rexistro de auditoría en `~/.cqlai_mcp_audit.log`. Os dous móvense a
`~/.cassandra` a próxima vez que se usan, así que un cliente configurado co
token antigo segue funcionando.

Quen teña o token pode usar o que o servidor permite: gárdao coma un
contrasinal e non o inclúas en nada que compartas. Para cambialo, borra o
ficheiro: o seguinte `cqlai mcp` crea un novo, e a configuración do cliente
necesita o token novo. Se alguén máis ca ti pode ler o ficheiro, CQLAI non o
usa e dío; `chmod 600 ~/.cassandra/cqlai_mcp_token` arránxao.

**Servilo noutro enderezo, con TLS.** Estes axustes están en MCP SERVER de
PREFERENCES, no bloque `mcp` de `cqlai.json`, e como opcións de `cqlai mcp`.
Aplícanse ao arrincar `cqlai mcp`.

| PREFERENCES | `cqlai.json` | Opción | Que fai |
|---|---|---|---|
| Listen on | `listen` | `--listen` | O enderezo no que serve. `127.0.0.1` se non se configura |
| Port | `port` | `--port` | O porto. `7845` se non se configura |
| Require token | `token` | `--token` | Cada petición ten que levar o token. Desactivado se non se configura |
| TLS certificate | `tlsCert` | `--tls-cert` | Serve HTTPS con este certificado (PEM) |
| TLS key | `tlsKey` | `--tls-key` | A clave privada do certificado (PEM) |
| TLS client CA | `tlsClientCA` | `--tls-client-ca` | Pide a cada cliente un certificado asinado por esta CA (PEM) |

En calquera enderezo que non sexa o desta máquina (`127.0.0.1`, `::1` ou
`localhost`), a rede pode chegar ao servidor. Entón ten que usar TLS, e o token
ou unha CA de cliente. Sen eles non arrinca: a barra de estado amosa
`MCP: OFF`, e `FILE > MCP SERVER` di por que. Por exemplo:

```json
{
  "mcp": {
    "listen": "0.0.0.0",
    "token": true,
    "tlsCert": "/etc/cqlai/server.pem",
    "tlsKey": "/etc/cqlai/server.key"
  }
}
```

**`cqlai mcp --headless`** non ten shell. É para un cliente que arrinca CQLAI e
fala con el por stdin e stdout:

```json
{
  "mcpServers": {
    "cassandra-prod": {
      "command": "cqlai",
      "args": ["mcp", "--headless", "--connection", "prod"]
    }
  }
}
```

`--connection` escolle unha conexión gardada polo nome; sen ela úsase a conexión
por defecto. O contrasinal é o gardado coa conexión, ou `CQLAI_PASSWORD`.
Arrinca haxa ou non clúster, e conéctase na primeira chamada.

| Opción | Que fai |
|---|---|
| `--headless` | Sen shell: MCP por stdin e stdout |
| `--port N` | O porto no que a shell serve MCP (por defecto: o axuste, ou 7845) |
| `--listen ENDEREZO` | O enderezo no que a shell serve MCP (por defecto: o axuste, ou 127.0.0.1) |
| `--token` | Cada petición ten que levar o token (por defecto: o axuste, ou desactivado) |
| `--tls-cert FICHEIRO`, `--tls-key FICHEIRO` | Serve HTTPS con este certificado e esta clave (PEM) |
| `--tls-client-ca FICHEIRO` | Pide a cada cliente un certificado asinado por esta CA (PEM) |
| `--connection NOME` | Headless: a conexión gardada que se usa |
| `--permit SELECT,DESCRIBE` | Só estas ordes, das que permiten os axustes |
| `--keyspaces ks1,ks2` | Só estes keyspaces son visibles, dos que permiten os axustes |
| `--max-rows N` | O tamaño de páxina, se é menor que o dos axustes |
| `--audit-log RUTA` | Onde vai o rexistro de auditoría; `-` desactívao |
| `--config-file RUTA` | O ficheiro de configuración que se le |

As opcións que din que pode facer o servidor só poden restrinxir o que
permiten os axustes. Nunca o amplían. `--port`, `--listen`, `--token` e as
opcións TLS din como se serve, e substitúen aos axustes.

### As ferramentas

| Ferramenta | Que fai |
|---|---|
| `connection_info` | A que está conectado o servidor e que pode facer alí. O modelo debería chamala primeiro. |
| `list_keyspaces` | Os keyspaces visibles, incluídos os virtuais |
| `list_tables` | As táboas dun keyspace, cada unha coa súa clave de partición e de clustering |
| `get_schema` | As claves e columnas dunha táboa |
| `fuzzy_search` | Keyspaces e táboas cuxo nome coincide cunha palabra: escrito igual, ou que soa igual, sen vogais ou letras, ou cunha letra mal (`hyto` atopa `hayato`). Di por que coincide cada un. |
| `describe` | A definición CQL dun keyspace, táboa, tipo, índice, vista, función ou agregado, tal como a imprime `DESCRIBE` |
| `query` | Executa un `SELECT` e devolve unha páxina de filas, cun token para a seguinte; todas as filas con `Auto fetch` |
| `trace_query` | Executa un `SELECT` con trazado e devolve a traza coas filas |
| `node_status` | A configuración do nodo, os seus pools de fíos, clientes conectados, cachés, compactacións ou uso de disco, desde as súas táboas virtuais (Cassandra 4.0 e posteriores). Os axustes con segredos agóchanse. |
| `table_size` | A estimación de Cassandra do número de particións dunha táboa e o seu tamaño medio |
| `list_roles` | Os roles, e os permisos dun; nunca os seus contrasinais |
| `propose_change` | Propón un cambio para que o executes ti. CQLAI nunca o executa. Ver abaixo. |

As catro primeiras son as mesmas ferramentas que usa a IA da vista `CHAT`. Unha
ferramenta só se ofrece se os axustes permiten o que necesita: sen `SELECT`,
non se ofrecen `query`, `trace_query`, `node_status` nin `table_size`.

O servidor tamén ofrece:

- **Recursos:** a definición de cada keyspace visible, e a de calquera táboa en
  `cql://schema/{keyspace}/{table}`, para que o cliente a lea e a xunte. Avísase
  ao cliente cando cambia o esquema, cámbiese onde se cambie.
- **Prompts:** `review_table`, que revisa a definición dunha táboa, e
  `diagnose_query`, que traza unha consulta e explica en que se vai o seu
  tempo. Usan as mesmas instrucións que `Alt+A` na shell.

### Cambios propostos

O servidor MCP nunca executa `INSERT`, `UPDATE`, `DELETE`, `BATCH`, `CREATE`,
`ALTER`, `DROP` nin `TRUNCATE`. Un modelo que quere un chama a `propose_change`
coa sentenza. CQLAI compróbaa - unha sentenza, táboas co seu keyspace, ningunha
agochada - e devolve, sen executala:

- a sentenza, lista para copiar;
- a conexión na que se executaría;
- o que fará: avisos escritos en CQLAI para cada tipo de sentenza, non a
  explicación do propio modelo. Por exemplo, que `DROP TABLE` borra os datos en
  todos os nodos e só se fai snapshot se `auto_snapshot` está activo; que borrar
  unha columna non se pode desfacer; que cambiar a replicación non move datos
  ata que reparas; que un `DELETE` sen a clave primaria completa borra unha
  partición enteira; que `INSERT` sobrescribe unha fila existente.

Indícase ao modelo que che amose a sentenza e todos os avisos.

En `cqlai mcp`, a proposta aparece tamén na shell: os avisos na Consola e a
sentenza no prompt, sen executar. Preme Enter para executala. A shell sempre
pregunta antes, con Cancelar seleccionado, diga o que diga o axuste
`Confirm changes`, porque ninguén a escribiu. Se a editas, é túa, e confírmase
coma calquera sentenza que escribas.

### As sentenzas rexeitadas devólvense

Cando os teus axustes rexeitan algo - `SELECT` non permitido, un keyspace ou
táboa agochados, unha columna agochada, un escaneo - a resposta non é só
"rexeitado". Dá o motivo e a sentenza exacta, e indica ao modelo que cha amose,
para que a executes ti se queres. Unha ferramenta que os teus axustes non
ofrecen fai o mesmo coa sentenza que executaría: `describe` devolve o seu
`DESCRIBE`, `list_roles` o seu `LIST ROLES`.

En `cqlai mcp`, a sentenza ponse tamén no prompt, sen executar, co motivo na
Consola, e confírmase antes de executarse, coma un cambio proposto.

Cun rexeitamento non volve nada do clúster: só a sentenza, que é a do propio
modelo. `GRANT`, `REVOKE`, os cambios de roles e usuarios, `USE`, e o que non
sexa unha sentenza que CQLAI poida ler, rexéitanse sen devolverse.

### Que pode facer o modelo

Configúrase na xanela `PREFERENCES`, en dúas seccións:

- **MCP SERVER**: os keyspaces que ve o modelo, as táboas que nunca ve, as
  columnas cuxos valores se agochan, se se permiten os escaneos, os límites e
  onde vai o rexistro de auditoría.
- **MCP SERVER - PERMITTED COMMANDS**: `SELECT`, `DESCRIBE` e `LIST`, cada unha
  permitida ou non. As tres están permitidas agás que digas o contrario. Nada
  que cambie datos ou esquema se pode permitir: só se propón.

`CONNECT` ten as mesmas dúas seccións para cada conexión gardada. O dunha
conexión só pode restrinxir o de `PREFERENCES`; unha orde que `PREFERENCES` non
permite aparece atenuada e non se pode marcar.

Os axustes gárdanse en `cqlai.json`, baixo `mcp`. Na shell aplícanse en canto
se garda `PREFERENCES` ou `CONNECT`: unha orde que desmarcas rexéitase desde a
seguinte chamada, sen reconectar. O porto aplícase cando `cqlai mcp` volve
arrincar, e un servidor headless le os seus axustes ao arrincar.

### Seguridade

Non se confía no modelo: pode equivocarse, e o texto que le do clúster pode
tentar dirixilo. Por iso cada regra aplícase en CQLAI, en cada chamada, pida o
que pida o modelo.

- **Unha sentenza por chamada.** `SELECT ...; DROP ...` rexéitase.
- **Só se executan as lecturas permitidas.** CQLAI clasifica a sentenza; o que
  non recoñece rexéitase.
- **Os cambios nunca se executan.** Propóñense, co que farán, para que os
  executes ti. `GRANT`, `REVOKE`, os cambios de roles e usuarios e `USE` nin
  sequera se propoñen.
- **As táboas nómeanse como `keyspace.táboa`.** Un nome sen keyspace rexéitase.
- **O agochado segue agochado.** Os keyspaces e táboas agochados non se listan,
  non se describen, non se atopan ao buscar e non se nomean nun erro.
  `system_auth`, que garda os hashes dos contrasinais, sempre está agochado.
- **Os keyspaces do sistema.** Sen lista de keyspaces, `system`,
  `system_schema`, `system_traces` e os demais keyspaces `system_` só son
  visibles se se marca `System keyspaces` en MCP SERVER de PREFERENCES
  (`"systemKeyspaces": true`). Cunha lista, inclúa os que queira.
  `node_status` le `system_views` en ambos os casos, agás que estea en deny ou
  falte nunha lista de keyspaces.
- **As táboas do sistema omiten o que se agocha.** `system_schema`,
  `system.size_estimates` e algunhas táboas virtuais teñen unha fila por cada
  keyspace e táboa. Se se agocha algo - con deny, unha lista de keyspaces, ou
  sen os keyspaces do sistema - omítese cada fila sobre algo agochado, e un
  `SELECT` sobre esas táboas ten que devolver `keyspace_name` (e o nome da
  táboa) co seu propio nome, para comprobar cada fila. Se non se agocha nada,
  lense tal cal. As filas que describen as táboas de `system_auth` mantéñense:
  son iguais en todos os clústeres. Os seus datos seguen agochados.
- **Columnas agochadas.** Os seus valores volven como `[redacted]`. Nunha táboa
  con columnas agochadas, un `SELECT` ten que usar `*` ou nomes de columna, sen
  `JSON`, alias nin funcións, e sen condicións sobre esas columnas.
- **Escaneos.** `ALLOW FILTERING` e os agregados entre particións rexéitanse
  agás que se permitan.
- **Páxinas.** Unha consulta devolve 100 filas cada vez agás que `Page size`
  diga outra cousa, ou o cliente pida outro número con `page_size`, que se usa
  tal cal. Con `Auto fetch` marcado (`"autoFetch": true`), unha consulta
  devolve todas as filas nunha chamada, lendo ela mesma as páxinas. Nunha
  táboa grande poden ser moitas.
- **Límites.** Unha chamada á vez, ata 60 por minuto. Cada páxina lida ten o
  tempo de espera da petición. Os valores longos recórtanse.
- **Sen credenciais.** Ningunha ferramenta recibe nin devolve un usuario ou un
  contrasinal.
- **Esta máquina agás que se configure outra, e nunca unha páxina web.** A
  shell serve MCP en `127.0.0.1` agás que `Listen on` diga outra cousa.
  Rexéitase unha petición dunha páxina web nun navegador. O token está
  desactivado agás que se active, e sen el calquera cousa nesta máquina pode
  usar o que o servidor permite. En calquera outro enderezo, esíxense TLS e o
  token ou un certificado de cliente.
- **Unha sesión propia.** O servidor usa unha sesión propia co clúster, así que
  as consultas do modelo nunca cambian a consistencia nin o trazado da túa
  shell.
- **Rexistro de auditoría.** Cada chamada escríbese en `~/.cassandra/cqlai_mcp_audit.log`,
  rexeitamentos incluídos, sen os valores das sentenzas e sen filas. Só ti
  podes lelo.

O control máis forte é o rol de Cassandra co que entra CQLAI. Dálle á conexión
do servidor MCP un rol propio que só poida ler o necesario:

```sql
CREATE ROLE mcp_reader WITH LOGIN = true AND PASSWORD = '...';
GRANT SELECT ON KEYSPACE shop TO mcp_reader;
GRANT SELECT ON KEYSPACE system_traces TO mcp_reader;  -- para trace_query
GRANT DESCRIBE ON ALL ROLES TO mcp_reader;              -- para LIST ROLES
```

O que CQLAI non pode controlar: todo o que devolve unha ferramenta chega ao
provedor do modelo. Usa os keyspaces visibles e as columnas agochadas para
deixar fóra o que non debe saír.

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
AUTOSAVE PARQUET 'results/';
SELECT * FROM large_table WHERE condition = true;
AUTOSAVE OFF;
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

### Saída JSON (AUTOSAVE JSON e --format json)

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

## Agradecementos

CQLAI baséase na fundación establecida por varios proxectos de código aberto, particularmente Apache Cassandra. Estendemos o noso sincero agradecemento á comunidade de Apache Cassandra polo seu excelente traballo e contribucións ao campo das bases de datos distribuídas.

Apache Cassandra é un sistema de xestión de bases de datos NoSQL de código aberto e gratuíto, distribuído, de almacén de columnas anchas, deseñado para manexar grandes cantidades de datos en moitos servidores commodity, proporcionando alta dispoñibilidade sen ningún punto único de falla.

### Recursos de Apache Cassandra

- **Sitio Web Oficial**: [cassandra.apache.org](https://cassandra.apache.org/)
- **Código Fonte**: Dispoñible en [GitHub](https://github.com/apache/cassandra) ou no repositorio Git de Apache en `gitbox.apache.org/repos/asf/cassandra.git`
- **Documentación**: Guías e referencias completas dispoñibles no [sitio web de Apache Cassandra](https://cassandra.apache.org/)

CQLAI incorpora e estende funcionalidades de varias ferramentas e utilidades de Cassandra, mellorándoas para proporcionar unha experiencia de terminal moderna e eficiente para desenvolvedores e DBAs de Cassandra.

Animamos aos usuarios a explorar e contribuír ao proxecto principal de Apache Cassandra, así como a proporcionar comentarios e suxestións para CQLAI a través das nosas páxinas de [discusións de GitHub](https://github.com/axonops/cqlai/discussions) e [problemas](https://github.com/axonops/cqlai/issues).

## Comunidade e Soporte

### Participa
- **Comparte Ideas**: Visita as nosas [Discusións de GitHub](https://github.com/axonops/cqlai/discussions) para propoñer novas funcións
- **Reporta Problemas**: Atopaches un erro? [Abre un problema](https://github.com/axonops/cqlai/issues/new/choose)
- **Contribúe**: Damos a benvida a pull requests! Consulta [CONTRIBUTING.md](CONTRIBUTING.md) para as pautas
- **Danos unha Estrela**: Se atopas útil CQLAI, por favor dálle unha estrela ao noso repositorio!

### Mantente Conectado
- **Sitio Web**: [axonops.com](https://axonops.com)
- **Contacto**: Visita o noso sitio web para opcións de soporte

## Licenza

Este proxecto está licenciado baixo a licenza Apache 2.0. Consulta o arquivo [LICENSE](LICENSE) para máis detalles.

As licenzas de dependencias de terceiros están dispoñibles no directorio [THIRD-PARTY-LICENSES](THIRD-PARTY-LICENSES/). Para rexenerar as atribucións de licenza, executa `make licenses`.

## Avisos Legais

*Este proxecto pode conter marcas rexistradas ou logotipos de proxectos, produtos ou servizos. O uso de marcas rexistradas ou logotipos de terceiros está suxeito ás políticas de ditos terceiros.*

- **AxonOps** é unha marca rexistrada de AxonOps Limited.
- **Apache**, **Apache Cassandra**, **Cassandra**, **Apache Spark**, **Spark**, **Apache TinkerPop**, **TinkerPop**, **Apache Kafka** e **Kafka** son marcas rexistradas ou marcas comerciais de Apache Software Foundation ou as súas subsidiarias en Canadá, Estados Unidos e/ou outros países.
- **DataStax** é unha marca rexistrada de DataStax, Inc. e as súas subsidiarias en Estados Unidos e/ou outros países.

---

<div align="center">
  <p>Feito con  polo equipo de <a href="https://axonops.com">AxonOps</a></p>
</div>
