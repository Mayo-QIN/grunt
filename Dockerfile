# Build a CentOS based system

# Use CentOS 7, but could be any Linux
FROM centos:7

MAINTAINER "Daniel Blezek" blezek.daniel@mayo.edu

# Create a user and do everything as that user
VOLUME /data

# Build grunt
RUN yum install -y golang git wget curl
ENV GOPATH=/root

# Copy local files into GOPATH
ADD .  $GOPATH/src/github.com/Mayo-QIN/grunt/
RUN go install github.com/Mayo-QIN/grunt

# Install files
RUN mkdir -p /grunt.d
RUN cp /root/bin/grunt /bin/grunt
COPY gruntfile.yml /gruntfile.yml

# Start grunt in /data with the given gruntfile
WORKDIR /data
CMD ["/bin/grunt", "/gruntfile.yml"]

# We expose port 9901 by default
EXPOSE 9901:9901
