#!/bin/bash

make build
./gochat.o -port :3100 -name peer1 -version 1.0 
