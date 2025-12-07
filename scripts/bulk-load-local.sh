#!/bin/bash

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)

cd PROJECT_ROOT

echo 'Build TSV->RDF parser'

go build -o ./bin/tsv2rdf ./cmd/tsv2rdf

echo Done

echo Parse data

time ./bin/tsv2rdf < ./data/source/cskg.tsv

echo Done

echo Chunk files

mkdir ./data/out/chunked -p 2> /dev/null

gsplit -n l/4 -d ./data/out/data.rdf ./data/out/chunked/chunk_ --additional-suffix=.rdf

echo Done

echo Cleanup

rm -rf out t zw tmp 2> /dev/null
kill -9 $(lsof -ti:5080) 2> /dev/null

echo Done

sleep 1

echo Run Dgraph Zero

dgraph=./bin/dgraph

$dgraph zero --bindall --my=localhost:5080 &

echo Done

echo Temporarily increasing max open files limit

ulimit -n 200000

echo Done

echo Running bulk load

time $dgraph bulk \
  -f ./data/out/chunked \
  -s ./schema.dql \
  --reduce_shards=4 \
  --map_shards=4 \
  --out ./out

echo Done

echo Run DB server
# ./bin/dgraph alpha --bindall --my=localhost:7080 --zero=localhost:5080 -p ./out/0/p -w ./out/0/w &