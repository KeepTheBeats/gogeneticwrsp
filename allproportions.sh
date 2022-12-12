#!/bin/env bash

CURRENT_DIR=$(cd $(dirname $0); pwd)
for i in 0 25 50 75 100
do
  j=$(echo | awk "{print $i/100}")
  echo "${CURRENT_DIR}/experiments/continuousexperiment/main.go $j"
  go run ${CURRENT_DIR}/experiments/continuousexperiment/main.go $j
  cp ${CURRENT_DIR}/experimenttools/*.csv ${CURRENT_DIR}/experimenttools/${i}/
done
