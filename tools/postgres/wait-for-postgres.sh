#!/bin/bash

while ! nc -z localhost "$1"; do sleep 1; done;
