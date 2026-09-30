#!/bin/bash

curl -X POST -H "Content-Type: application/json" -d '@device_sample.json' localhost:8081/api/devices
