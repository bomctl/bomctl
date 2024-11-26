FROM gcr.io/oss-fuzz-base/base-builder-go
RUN apt-get update && apt-get install -y make autoconf automake libtool
COPY . $SRC/bomctl_dir
COPY .clusterfuzzlite/build.sh $SRC/build.sh
WORKDIR $SRC/bomctl_dir
RUN git clone https://github.com/bomctl/bomctl
