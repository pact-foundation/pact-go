#!/bin/bash

goveralls -service="travis-ci.com" -coverprofile=profile.cov -repotoken $COVERALLS_TOKEN