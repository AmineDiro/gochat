#!/bin/bash

make build
./gochat.o -port :3100 -name Bob -version 1.0  
