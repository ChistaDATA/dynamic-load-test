#!/bin/bash

file=$1

while IFS= read -r line
do
  echo "Triggering worker for task file: $line"
  TASK_FILE_NAME=$line locust --worker -f locustfile.py &
done < "$file"

wait
