
FROM pesscara/grunt

USER root
RUN yum install -y wget

USER grunt
# RUN wget https://github.com/stnava/ANTs/releases/download/v2.1.0/Linux_X_64.tar.bz2.RedHat
# RUN tar fxvj Linux_X_64.tar.bz2.RedHat
COPY ants/ ants/
COPY ants.gruntfile.yml gruntfile.yml
COPY simpleReg simpleReg

