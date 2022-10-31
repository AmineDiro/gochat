#!/bin/bash

make build
./gochat -port :3100 -name peer1 -version 1.0 -v 
