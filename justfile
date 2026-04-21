set shell := ["zsh", "-cu"]

buf-lint:
    GOCACHE=/tmp/gocache go run github.com/bufbuild/buf/cmd/buf@v1.59.0 lint

buf-breaking:
    GOCACHE=/tmp/gocache go run github.com/bufbuild/buf/cmd/buf@v1.59.0 breaking --against '.git#branch=main'

proto-clean:
    rm -f auth/v1/*.pb.go registration/v1/*.pb.go user/v1/*.pb.go

proto-gen:
    just proto-clean
    GOCACHE=/tmp/gocache go run github.com/bufbuild/buf/cmd/buf@v1.59.0 generate
