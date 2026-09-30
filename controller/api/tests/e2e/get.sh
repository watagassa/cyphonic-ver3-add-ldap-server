#!/bin/bash

# Healthz
curl -i 'localhost:8081/healthz'

echo "GetDevices()"
curl -s 'localhost:8081/api/device?device_id=61df7869c6ea4fca675a6272d25e80cb841d0c4ce1b2b2b7e0815bb2' | jq

echo "GetDevices()"
curl -s 'localhost:8081/api/devices' | jq

echo "GetGeneralDevice()"
curl -s 'localhost:8081/api/general_device?device_id=61df7869c6ea4fca675a6272d25e80cb841d0c4ce1b2b2b7e0815bb2' | jq

echo "GetGeneralDevices()"
curl -s 'localhost:8081/api/general_devices' | jq

echo "GetAccounts()"
curl -s 'localhost:8081/api/accounts?account_id=1' | jq

echo "GetChildDeviceInformations()"
curl -s 'localhost:8081/api/adapter/2bac8cfcfa0a0e8fe7655cff0fec896866eff52e94efce030be15337/children' | jq
