#!/bin/bash
make build
./gochat.o -port :3200 -name peer2 -version 1.0 -v -peers :3100
