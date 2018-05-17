#!/bin/bash

goveralls -service="travis-ci" -coverprofile=profile.cov -repotoken "${COVERALLS_TOKEN}"