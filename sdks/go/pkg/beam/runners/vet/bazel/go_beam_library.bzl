load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_context",
    "go_library",
)
load(
    ":beam_pipeline.bzl",
    "go_beam_code_generator_binary",
)

def go_beam_library(name, srcs, package, importpath, **kwargs):
    # 1) Create a library with everything but the generated srcs.
    go_library_attrs = dict(kwargs)
    go_library_attrs["importpath"] = importpath
    go_library_attrs["srcs"] = srcs

    private_library_name = name + "_without_generated_code"
    go_library(
        name = private_library_name,
        **go_library_attrs
    )

    generated_src = name + "_generated.go"

    # 2) Run the beam pipeline in code generation mode using the library from 1.
    go_beam_code_generator_binary(
        name = name + "_code_generator",
        out = generated_src,
        generator_deps = [
            ":" + private_library_name,
        ],
        generator_src_out = name + "_code_generator.go",
        package = package,
        pipeline_importpath = importpath,
        visibility = ["//visibility:private"],
    )

    # 3) Create a final go_library with srcs + the generated .go file.
    final_go_library_attrs = dict(go_library_attrs)
    final_go_library_attrs.pop("srcs")
    final_go_library_attrs.pop("deps")
    go_library(
        name = name,
        srcs = [generated_src],
        deps = [
            "//go/pkg/beam",
            "//go/pkg/beam/runners/vet",
            "//go/pkg/beam/core/util/reflectx:reflectx",
            "//go/pkg/beam/core/runtime:runtime",
        ],
        embed = [private_library_name],
        **final_go_library_attrs
    )
