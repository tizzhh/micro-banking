#!/bin/bash

echo "make migrations"

source env/goose.env
goose up

./bank_service