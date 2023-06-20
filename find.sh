#!/bin/bash
for file in $(find "./dist" -type f | grep -E '\.(gz|zip)$'); do 
 echo "Processing file: $file"
 filename=$(basename "$file")
 echo "$filename"
done
