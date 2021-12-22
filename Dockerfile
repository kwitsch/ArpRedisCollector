FROM ghcr.io/kwitsch/docker-buildimage:main AS build-env

ADD src .
RUN gobuild.sh -o arprediscollector

FROM scratch
COPY --from=build-env /builddir/arprediscollector /arprediscollector

ENTRYPOINT ["/arprediscollector"]