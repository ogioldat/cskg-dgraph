# Project ADB

## Database propositions

- dgraph,
- https://cayley.gitbook.io/cayley/

## Checklist

- [x] Choice of technology.
- [x] Architecture: components and interactions, optional diagram.
- [x] Prerequisites (software modules, databases, etc.).
- [x] Installation and setup instructions.
- [x] Design and implementation process, step by step.
- [x] Details on how each of the goals is addressed, including database queries and logic behind them.
- [x] The roles of all the students in the project and description of who did what.
- [x] Results, including example runs, outcomes and timings indicating efficiency.
- [x] User manual (how to run the software) and a step-by-step manual how to reproduce the results.
- [x] Self-evaluation: efficiency should be discussed, strategies for future mitigation of identified shortcomings.

## Data overview

* `id`: Row identifier.
* `node1`: Node 1 identifier.
* `relation`: Edge/relation type.
* `node2`: Node 2 identifier.
* `node1;label`: Human-readable label for node 1.
* `node2;label`: Human-readable label for node 2.
* `relation;label`: Human-readable label for the edge/relation.
* `relation;dimension`: Unused field.
* `source`: Record source.
* `sentence`: Sentence the term is used in.

There are 6001531 records.

## Running setup

### Option 1 (Recommended)

Simplest option to run the setup, uses backup file from Google Drive.

1. `cd ci`
2. `./scripts/download-data.sh`
3. `docker compose up dgraph-restore`
4. `docker compose up dgraph`
5. `docker compose up dgraph-client`

### Option 2

Inserts RDF files with bulk loader.

1. Requires to download CSKG, must be saved to `data/source/cskg.tsv`.
2. Build `cms/tsv2rdf`.
3. Run `tsv2rdf` binary for `cskg.tsv`.
4. `docker compose up`.
