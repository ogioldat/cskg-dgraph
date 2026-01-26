# Project ADS:

This project implements a graph database solution for the Common Sense Knowledge Graph (CSKG) using **Dgraph**. It includes a complete CI/CD pipeline for automated performance benchmarking, a custom Go client for complex traversals, and a pre-computed schema optimization strategy.

## 1. Choice of Technology

* **Database:** [Dgraph](https://dgraph.io/) - Chosen for its native graph capabilities and horizontal scalability.
* **Client Implementation:** Go (Golang) - Selected for high performance and concurrency support (goroutines) during complex graph traversals.
* **Infrastructure:** Docker & Docker Compose - Used for consistent environment setup and "Clean State" benchmarking.
* **Analysis:** Python (Pandas) - Used for processing benchmark logs.

## 2. Architecture

The system consists of three main containerized components:

1. **Dgraph Alpha (Server):** The core database engine storing the CSKG data.
2. **Dgraph Restore (Init):** A transient container that downloads a backup and populates the database volume before the server starts. This ensures every test run starts with an identical, clean state.
3. **Dgraph Client (Benchmarker):** A custom Go application (`cmd/client`) that resides in the same network. It executes queries and measures latency from within the cluster to minimize network noise.

## 3. Data Overview

The dataset is the **Common Sense Knowledge Graph (CSKG)**.

* **Records:** ~6,001,531 edges.
* **Schema Fields:**
  * `node1`, `node2`: Unique identifiers (URIs).
  * `relation`: The edge type (e.g., `/r/Synonym`).
  * `label`: Human-readable text.
  * `sentence`: Context usage.

**Pre-computation:** To overcome Dgraph's limitations with "zero-count" queries, we pre-calculated `ingoing_cnt`, `outgoing_cnt`, and `neighbors_cnt` during the ingestion phase and stored them as integer predicates on the nodes.

## 4. Design & Implementation

### Data Ingestion Strategy

Instead of using live loading, we implemented a custom **TSV-to-RDF converter** (`cmd/tsv2rdf`).

1. Parses the raw CSKG CSV.
2. Calculates edge counts in memory.
3. Generates a Dgraph-compatible RDF file with explicit count predicates.
4. Creates a binary backup for fast restoration in CI.

### Schema & Indexing

We utilize Dgraph's advanced indexing to support the queries:

* `uri`: `@index(exact)` for O(1) lookups.
* `label`: `@index(term, exact)` for text search.
* `rel`: `@reverse @count` to allow bidirectional traversal and native counting.

## 5. Goals & Queries

The system addresses 5 specific query patterns, ranging from simple lookups to complex graph algorithms:

1. **Query 1 (Point Lookup):** Fetch all details for a specific Node URI.
2. **Query 9 (Aggregation):** Count total concept nodes.
3. **Query 10 (Zero Successors):** Find nodes with `outgoing_cnt == 0`. *Optimized using pre-computed fields.*
4. **Query 12 (Most Neighbors):** Find top 10 nodes by neighbor count. *Optimized using pre-computed fields.*
5. **Query 17 (Distant Synonyms):** Find "distant synonyms" (e.g., synonym of an antonym of an antonym).
   * **Logic:** Implemented as a **Client-Side BFS** in Go.
   * **Reasoning:** Dgraph's native query language (DQL) could not track the dynamic "Sign" state (positive/negative) required to distinguish synonyms from antonyms during recursion.

## 6. Installation & Setup

### Prerequisites

* Docker and Docker Compose.
* Go 1.19+ (optional, for local development).

### Option 1: Fast Start (Recommended)

Uses a pre-built backup from Google Drive.

1. Navigate to the CI folder:
   ```bash
   cd ci
   ```
2. Download the backup files:
   ```bash
   ./scripts/download-backup.sh
   ```
3. Start the restoration container (loads data to volume):
   ```bash
   docker compose up dgraph-restore
   ```
4. Start the database and client:
   ```bash
   docker compose up -d dgraph dgraph-client
   ```

### Option 2: Build from Source

Inserts RDF files using the bulk loader.

1. Download CSKG to `data/source/cskg.tsv`.
2. Build the converter: `go build -o tsv2rdf cmd/tsv2rdf/main.go`
3. Convert data: `./tsv2rdf`
4. Run `docker compose up`.

## 7. User Manual: Reproducing Results

To run the full performance benchmark (as done in CI):

1. Ensure the stack is running (Option 1 above).
2. Execute the test script:
   ```bash
   ./ci/scripts/test.sh 5
   ```

   *(The argument `5` specifies the number of iterations).*
3. Logs will be generated in `ci/logs/`.

## 8. Results

Performance metrics averaged over 5 iterations on standard CI hardware:

| Goal         | Description           | Duration (ms) | Client Mem (MB) | DB Mem (MB) | Calculation      |
| :----------- | :-------------------- | :------------ | :-------------- | :---------- | :--------------- |
| **1**  | Simple Lookup (Batch) | 51800         | 6.8             | 4956.0      | Full DB          |
| **9**  | Total Count           | 1000          | 5.3             | 5146.4      | Full DB          |
| **10** | Zero Successors       | 1000          | 5.8             | 5042.7      | Full DB          |
| **12** | Most Neighbors        | 1000          | 9.8             | 6996.6      | Full DB          |
| **17** | Distant Synonyms      | 2000          | 56.1            | 5950.6      | Partial (Client) |

*Note: Goals 9, 10, and 12 are near-instant due to the pre-computation strategy. Goal 17 shows higher Client Memory usage because the graph traversal logic resides in the Go application.*

## 9. Self-Evaluation

**Efficiency:**
The database performs well for indexed lookups and pre-computed aggregations (Goals 1-12).

**Shortcomings & Mitigation:**

* **Complex Recursion:** Dgraph struggles with stateful recursion (Query 17). We mitigated this by moving logic to the "Thick Client," effectively using Dgraph as a storage engine and Go as the compute engine.
* **Zero-Counts:** Dgraph does not index "missing" edges. We mitigated this by materializing counts into the schema.

## 10. Student Roles

#### Tomasz Ogiołda

* **Architecture Design:** Designed the containerized setup using Docker/Podman, including the "restore-first" strategy for clean-state benchmarking.
* **Data Pipeline:** Implemented the `tsv2rdf` tool (Go) to convert the raw CSKG dataset into Dgraph-compatible RDF format, including the pre-computation logic for edge counts.
* **Complex Algorithms:** Implemented the client-side Breadth-First Search (BFS) logic for "Distant Synonym" traversal (Query 17) to overcome DQL limitations.
* **CI/CD Automation:** Configured the GitLab CI pipeline (`.gitlab-ci.yml`) and created the shell scripts for automated backup retrieval and environment orchestration.
* **Benchmarking & Analysis:** Wrote the testing scripts (`test.sh`), executed the performance experiments, and performed the data analysis on the resulting logs.

#### Szymon Pająk:

* **Go Client Implementation:** Developed the custom Go application (`cmd/client`) to interface with Dgraph, handling connection pooling and query execution.
* **Schema & Indexing:** Designed the Dgraph schema (`.dql`) and optimized indices for performance.
* **Query Formulation:** Formulated and optimized all GraphQL/DQL queries for the defined project goals.
* **Benchmarking & Analysis:** Wrote the testing scripts (`test.sh`), executed the performance experiments, and performed the data analysis on the resulting logs.
