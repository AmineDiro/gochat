#!/bin/bash
make build
./gochat.o -port :3200 -name peer2 -version 0.2 -v -peers :3100
