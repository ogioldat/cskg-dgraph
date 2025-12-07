#!/bin/bash

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)

cd PROJECT_ROOT

mkdir -p source

curl -L "https://zenodo.org/records/4331372/files/cskg.tsv.gz?download=1" -o source/cskg.tsv.gz