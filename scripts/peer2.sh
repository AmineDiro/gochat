#!/bin/bash
make build
./gochat.o -port :3200 -name Alice -version 1.0  -peers :3100
