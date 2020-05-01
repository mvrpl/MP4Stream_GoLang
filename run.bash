#!/usr/bin/env bash

go build -o streamApp main.go && ./streamApp -v "./videos/sample_720p.mp4"