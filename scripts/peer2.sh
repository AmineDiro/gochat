#!/bin/bash
make build
./gochat -port :3200 -name peer2 -version 1.0 -v -peers :3100
