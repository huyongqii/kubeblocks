#!/bin/bash

for file in $(find "./dist" -type f | grep -E '\.(deb|rpm)$'); do
  echo "Processing file: $file"
done
