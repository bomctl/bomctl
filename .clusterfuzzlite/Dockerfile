FROM gcr.io/oss-fuzz-base/base-builder-go
RUN apt-get update && apt-get install -y make autoconf automake libtool
RUN git clone https://github.com/bomctl/bomctl
COPY . $SRC/bomctl
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/bomctl
