#!/bin/bash

echo "Downloading backup data from Google Drive..."

mkdir -p ./data/backup

curl -L "https://github.com/ogioldat/cskg-dgraph/releases/download/backup-1/backup-20260117T205507Z-1-001.zip" -o ./data/backup.zip

unzip ./data/backup.zip -d ./data/

echo "Backup data downloaded."
