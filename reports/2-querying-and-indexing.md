# Querying and indexing report

## Problems

### DB crashes on unrestricted query timeouts

Whenever we ran a query over Ratel that does not have any timeout restrictions ($\infty$ by default), such procedure caused the DB container to crash...

```log
dgraph-zero      | E1129 19:02:08.876881       1 pool.go:303] CONN: Unable to connect with alpha:7080 : rpc error: code = Unavailable desc = name resolver error: produced zero addresses
dgraph-zero      | E1129 19:02:09.930873       1 pool.go:303] CONN: Unable to connect with alpha:7080 : rpc error: code = Unavailable desc = name resolver error: produced zero addresses
dgraph-zero      | E1129 19:02:10.993449       1 pool.go:303] CONN: Unable to connect with alpha:7080 : rpc error: code = Unavailable desc = name resolver error: produced zero addresses
dgraph-zero      | E1129 19:02:12.039417       1 pool.go:303] CONN: Unable to connect with alpha:7080 : rpc error: code = Unavailable desc = name resolver error: produced zero addresses
```

We had to recreate the environment.

### Querying limitations

Dgraph's data access pattern was not designed for complex queries... It excels with graph subpart retrieval, usually by `uid`, for more complex queriers it will choke and timeout.

#### Zero counts

Dgraph doesn't track 0 counts, which create a huge problem for some of the queries. For example, the query below will fail with `count(predicate) cannot be used to search for negative counts (nonsensical) or zero counts (not tracked)`.

```gql
query no_successors {
  q(func: eq(count(rel), 0)) {
    uid
    label
  }
}
```

An alternative to querying all nodes with no successors would be:

```gql
query {
  q(func: type(Concept)) @filter(eq(count(rel), 0)) {
    uid
    label
  }
}
```

**However**, the query above fetches all nodes and then filters out not matching result, in reality this means sequential scan over >2M nodes. It caused timeouts.

To overcome the limitations, we precomputed ingoing, outgoing and neighbor edge counts for each node.

#### Context aware queries

Queries such as:

> The dataset contains many synonyms and antonyms (their ids are "/r/Synonym", "/r/Antonym"). Let's define a distant synonym and antonym as a node connected by a series of synonym or antonym relations. Synonym of a synonym is a synonym, synonym of an antonym is an antonym, antonym of an antonym is a synonym, etc. Find all distant synonyms of a given node at a specified distance (length of the shortest path), given as a command parameter.

Are can't be modeled in DQL without external reasoning and context tracking engine.

#### Shortest path inefficiency

Shortest-path (`@shortest()`) queries are inefficient because it performs unconstrained BFS over a massive, densely connected graph, causing exponential state expansion that hits hundreds of thousands of nodes even at modest depths.

## Decisions

How the problems were overcome.

### Indexing

Each predicate is indexed

```gql
uri: string @index(exact) .
label: string @index(term, exact) .
rel: [uid] @reverse @count .
```

1. `uri` and `label` uses index for `exact` match allowing for the following ops `le, ge, lt, gt, eq`.
2. `rel`, which represents relation predicate `uid`s must have indexed `@count` for aggregative queries, as well as `@reverse` clause to enable querying for node parents (`~rel` parents, `rel` children).

### Precompute edge counts

The schema was adjusted accordingly, the edge count predicates are indexed for to use `le, ge, lt, gt, eq`.

```gql
uri: string @index(exact) .
label: string @index(term, exact) .
rel: [uid] @reverse @count .
ingoing_cnt: int @index(int) .
outgoing_cnt: int @index(int) .
neighbors_cnt: int @index(int) .

type Concept {
    uri
    label
    rel
    ingoing_cnt
    outgoing_cnt
    neighbors_cnt
}
```

Out Golang program for TSV->RDF parsing now does one more thing -- aggregate counts of edges for each node, write the data at the end (not to create redundant RDFs).

```go
for _, currNode := range nodeCache {
 fmt.Fprintf(buf, "\n")
 fmt.Fprintf(buf, `<_:%s> <ingoing_cnt> "%d"^^<xs:int> .`, currNode.nodeId, currNode.ingoingCnt)
 fmt.Fprintf(buf, "\n")
 fmt.Fprintf(buf, `<_:%s> <outgoing_cnt> "%d"^^<xs:int> .`, currNode.nodeId, currNode.outgoingCnt)
 fmt.Fprintf(buf, "\n")
 fmt.Fprintf(buf, `<_:%s> <neighbors_cnt> "%d"^^<xs:int> .`, currNode.nodeId, currNode.outgoingCnt+currNode.ingoingCnt)
 }
```

This solved the problem, query is in the following form:

```gql
query no_predecessors($first: int = 1000, $offset: int = 0, $id: int) {
  var(func: eq(ingoing_cnt, 0)) {
    S as uid
  }

  total(func: uid(S)) {
    count(uid)
  }

  q(func: uid(S), first: $first, offset: $offset) {
    uid
    label
    ~rel {
        uid
    }
  }
}
```
