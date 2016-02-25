# Build a CentOS based system

FROM pesscara/grunt

MAINTAINER "Daniel Blezek" blezek.daniel@mayo.edu

# Install Slicer 4.4
ADD Slicer-4.4.0-linux-amd64.tar.gz /grunt
COPY slicer.gruntfile.yml gruntfile.yml

