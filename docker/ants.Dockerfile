FROM pesscara/grunt
USER root
# Maybe I need to install Dev tools for dependancies
RUN yum install -y wget
RUN yum install -y git
RUN yum install zlib-devel bzip2-devel openssl-devel ncurses-devel sqlite-devel readline-devel tk-devel -y
RUN yum install libitpp atlas blas lapack atlas-devel blas-devel lapack-devel libpng-devel -y
RUN yum groupinstall "Development tools" -y
WORKDIR /tmp/
RUN git clone -b release http://cmake.org/cmake.git
RUN mkdir /tmp/cmake-build
WORKDIR /tmp/cmake-build
RUN ../cmake/bootstrap
RUN make -j4
RUN ./bin/cmake -DCMAKE_BUILD_TYPE:STRING=Release .
RUN make
RUN make install
# Install ANTS
WORKDIR /tmp/
RUN git clone https://github.com/stnava/ANTs.git
RUN mkdir -p /tmp/build
RUN cd ./build/
RUN cmake ./ANTs -DCMAKE_BUILD_TYPE=Release  -DBUILD_EXAMPLES=OFF -DBUILD_TESTING=OFF
RUN make -j8
RUN echo export PATH=/tmp/build/bin:\$PATH >> ~/.bashrc
RUN echo export ANTSPATH=${ANTSPATH:="/tmp/build/bin"} >> ~/.bashrc
# copy .yml file as well te script to run. Need to modify so it works.
USER grunt
WORKDIR /grunt
COPY docker/ants.gruntfile.yml /grunt/gruntfile.yml
COPY simpleReg simpleReg