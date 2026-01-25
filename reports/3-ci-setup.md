# CI Setup

## Overview

- we use dgraph backup feature and host file publicly on google drive (simple and efficient solution),
- ci downloads the backup,
- we run dgraph restore (admin feature) to load backup,

```bash
dgraph restore \
    --location /backup \
    --force_zero=false \
    --postings /dgraph
```

- then we run database, backup replaces DBs files with data,
- our DB is using go, we have a `client.Dockerfile` that compiled go binary (it preloads all possible gql queries and provides db connection to dgraph),
- to make performance measures reliable, we wanted to run client for the whole benchmark in one container, thus we put dummy infinite sleep to the client container to keep it alive and running, this let's us do podman execs to the container,
- in our `test.sh` script we measure container resource consumption during task execution:
  - first we start streaming podman stats for both containers, we gather theirs PIDs to kill them after script completes,
  - first and last log row in container logs indicate start and end time of the experiment,
  - we execute
    - task 1 for each node in data/sample-nodes.csv,
    - tasks 12,9,10,17 are executed as discussed
  - experiments are repeated 5 times,
  - final results are in artifacts `logs` folder,

## Tasks

### Query 1

Simple get by uri field.

```gql
query all($id: string, $first: int = 10000, $offset: int = 0, $uri: string) {
  q(func: eq(uri, $uri)) {
    uid
    uri
    label
    rel @facets(label, edge_id) {
      uid
      uri
      label
    }
  }
}
```

### Query 9

Use aggregate fn.

```gql
query count_all_nodes {
  node_count(func: type(Concept)) {
    count(uid)
  }
}
```

### Query 10

We had to precompute outgoing and ingoing edges for each node, without it querying performance was terrible.

```gql
query no_successors($first: int = 10000, $offset: int = 0, $id: int) {
  var(func: eq(outgoing_cnt, 0)) {
    S as uid
  }

  total(func: uid(S)) {
    count(uid)
  }

  q(func: uid(S), first: $first, offset: $offset) {
    uid
    label
    rel {
      uid
    }
  }
}
```

### Query 12

Similarly here, precomputed.

```gql
query most_neighbors {
  q(func: type(Concept), orderdesc: neighbors_cnt, first: 10) {
    uid
    label
  	rel @facets {
      uid
      label
    }
  	outgoing_cnt
  	ingoing_cnt
	neighbors_cnt
  }
}
```

### Query 17*

We had to implement a graph traversal algorithm on the client's side due to dgraph's querying limitations.
