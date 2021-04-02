FROM golang AS build

WORKDIR /build
ADD . /build/.

RUN go build -o avahi-sync

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=build /build/avahi-sync .

ENTRYPOINT [ "./avahi-sync" ]
CMD [] 