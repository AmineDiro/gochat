#!/bin/bash
make build
./gochat.o -port :3300 -name peer3 -version 1.0 -v -peers :3100
