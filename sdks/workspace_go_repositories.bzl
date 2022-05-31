load("@bazel_gazelle//:deps.bzl", "go_repository")

def gazelle_managed_go_repositories():
    """Repositories added by gazelle."""

    # See https://github.com/bazelbuild/rules_go/blob/4a42b4092abdc60d14419a79afaec3659fbceb26/go/workspace.rst#id8
    go_repository(
        name = "org_golang_google_grpc",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/grpc",
        sum = "h1:u+MLGgVf7vRdjEYZ8wDFhAVNmhkbJ5hmrA1LMWK1CAQ=",
        version = "v1.46.2",
    )
