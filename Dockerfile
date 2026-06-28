FROM gcr.io/distroless/static-debian12:nonroot
COPY octane /octane
ENTRYPOINT ["/octane"]
