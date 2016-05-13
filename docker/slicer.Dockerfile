# Build a CentOS based system

FROM pesscara/grunt

MAINTAINER "Daniel Blezek" blezek.daniel@mayo.edu

# Install Slicer 4.5
# Dumb kitware can't publish normal URLs like the rest of us...
# this downloads slicer as /download
# ADD http://slicer.kitware.com/midas3/download?bitstream=461634 /
ADD docker/Slicer-4.5.0-1-linux-amd64.tar.gz /
# CMD mv /download /Slicer-3.5.0-1-linux-amd64.tar.gz
CMD tar fx /Slicer-4.5.0-1-linux-amd64.tar.gz
CMD rm /Slicer-4.5.0-1-linux-amd64.tar.gz
CMD mv /Slicer-4.5.0-1-linux-amd64 /Slicer

# Add the grunt config
ADD docker/slicer.gruntfile.yml /grunt.d/slicer.gruntfile.yml
