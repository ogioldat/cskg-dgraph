# Data ingestion report

## Data migration options

Dgraph supports a couple of data ingestion strategies.

### Live Import (Live Loader)

Designed for loading data into a loading cluster.

### Initial Import (Bulk Loader)

Only suitable when creating cluster from scratch, directly builds SS Tables of the database. Great parallelization configuration options. Suitable for substantial datasets.

### Import CSV Data

CSV, but not really. It is required to parse the CSV file into the JSON array, documentation suggests jq -- not suitable for big data volumes.

### Import MySQL Data

Import data from MySQL -> Dgraph.

## Approach

Out primary choice for data ingestion was `Initial Import`. However we experienced issues with this approach. The schema was not getting loaded correctly -- always only one field from schema was loaded, the others were ignored. Data was not getting inserted, just `uid` + just one of the schema fields that managed to load.

We decided to use `Live Import` and managed to load the data. There we couple problems on the way, such as db timeouts. To mitigate the timeouts we split data into small chunks 1000 entries per JSON file and 7 separate file directories -- not to load all at once.
