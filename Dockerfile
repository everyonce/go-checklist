FROM centurylink/ca-certs
MAINTAINER Daniel Ness <daniel@everyonce.com>
WORKDIR /app
COPY go-checklist /app/
ENV PORT 8080
EXPOSE 8080
ENTRYPOINT ["./go-checklist"]
