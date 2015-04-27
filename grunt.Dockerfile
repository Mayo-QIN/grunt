# Build a CentOS based system

# Use CentOS 7, but could be any Linux
FROM centos:7

MAINTAINER "Daniel Blezek" blezek.daniel@mayo.edu

# Install files
COPY bin/grunt /grunt

# What do we run on startup?
ENTRYPOINT /grunt
