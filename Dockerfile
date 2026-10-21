# Release image, built by goreleaser (see .goreleaser.yaml `dockers`).
#
# goreleaser places the cross-compiled `octane` binary and the files listed
# in `extra_files` at the root of the build context.
#
# The image serves three uses, so it sets no ENTRYPOINT:
#   - GitHub Action: action/action.yml runs /entrypoint.sh.
#   - GitLab CI and other runners: `script:` lines call `octane` from PATH.
#   - docker run: `docker run --rm <image> octane --version`.

FROM alpine:3.23

RUN apk add --no-cache bash ca-certificates

COPY octane /usr/local/bin/octane
COPY --chmod=0755 action/entrypoint.sh /entrypoint.sh

CMD ["octane", "--help"]
