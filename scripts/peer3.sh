#!/bin/bash
make build
./gochat -port :3300 -name peer3 -version 1.0 -v -peers :3100
