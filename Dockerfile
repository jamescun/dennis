# =========================================================================== #
# Rootless containers for DENNIS.
# =========================================================================== #
# STEP 1: Builder
#
# Build Arguments:
#   VERSION  semantic release of DENNIS from git tag.
#   COMMIT   git commit of source code when building.
# --------------------------------------------------------------------------- #
# Builder contains the Go compiler, build related utilities and system files
# necessary to build a DENNIS, to be consumed by the result stage.
# --------------------------------------------------------------------------- #
FROM golang:1.26 AS builder

# Base directory where DENNIS will be copied to and built from.
WORKDIR /go/src/github.com/jamescun/dennis

# VERSION and COMMIT are build arguments that are injected into the DENNIS
# binary at compile time.
ARG VERSION="0.0.0"
ARG COMMIT="main"

# Initialize DENNIS-specific directories to be copied later.
RUN mkdir /data

# Copy the go modules definitions first to take advantage of layer caching
# between runs when nothing has changed.
COPY go.mod go.sum ./
RUN go mod download

# Finally copy the source of DENNIS to compile.
COPY . .

# Compile the DENNIS binary, embedding the version/commit at the time of
# build.
RUN CGO_ENABLED=0 go build -tags package -trimpath -o /bin/dennis \
	-ldflags "-s -w -X github.com/jamescun/dennis/app/pkg/build.version=${VERSION} -X github.com/jamescun/dennis/app/pkg/build.commit=${COMMIT}" \
	main.go


# --------------------------------------------------------------------------- #
# STEP 2: Result
# --------------------------------------------------------------------------- #
# Result is a rootless container containing only DENNIS and a minimal set
# of supporting files.
# --------------------------------------------------------------------------- #
FROM scratch

# Copy files expected by the containerizer and various Go libraries of a
# minimum viable Linux filesystem.
COPY extra/passwd /etc/passwd
COPY extra/group /etc/group
COPY config.example.yml /etc/dennis/config.yml
COPY --from=builder --chown=65534:65534 /data /data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Configure nobody:nobody as the default user/group, this can be overwritten
# with `--user uid:gid` given when starting the container.
USER 65534:65534
ENV UID=65534 GID=65534 USER=nobody GROUP=nobody

# Copy DENNIS binary from builder.
COPY --from=builder /bin/dennis /bin/dennis

# Configure /bin/dennis as the default binary executed when running the
# container.
ENTRYPOINT [ "/bin/dennis" ]
