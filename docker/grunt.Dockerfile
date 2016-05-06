# Build a CentOS based system

# Use CentOS 7, but could be any Linux
FROM centos:7

MAINTAINER "Daniel Blezek" blezek.daniel@mayo.edu

# Create a user and do everything as that user
VOLUME /data

# Install files
RUN mkdir -p /grunt.d
COPY bin/grunt-docker /bin/grunt
COPY docker/gruntfile.yml /gruntfile.yml

# Start grunt in /data with the given gruntfile
WORKDIR /data
CMD ["/bin/grunt", "/gruntfile.yml"]

# We expose port 9901 by default
EXPOSE 9901:9901
