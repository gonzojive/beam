load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_context",
)

def _go_beam_code_generator_binary(ctx):
    go = go_context(ctx)
    ctx.actions.run(
        inputs = [],
        outputs = [ctx.outputs.out],
        arguments = [
            "--output",
            ctx.outputs.out.path,
            "--template_json",
            json.encode_indent(struct(
                import_path = ctx.attr.blank_import,
            )),
        ],
        progress_message = "Generating program that will generate code for package %s" % ctx.attr.blank_import,
        executable = ctx.executable._main_file_generator_tool,
    )
    return [DefaultInfo()]

# See https://bazel.build/rules/rules-tutorial and
# https://github.com/bazelbuild/rules_go/blob/master/go/toolchains.rst#writing-new-go-rules
# for information about writing custom go toolchains.
#
# Examples:
# https://sourcegraph.com/github.com/bazelbuild/rules_go/-/blob/go/private/rules/binary.bzl
go_beam_code_generator_binary = rule(
    implementation = _go_beam_code_generator_binary,
    attrs = {
        "out": attr.output(mandatory = True),
        "blank_import": attr.string(),
        "_go_config": attr.label(default = "@io_bazel_rules_go//:go_config"),
        "_stdlib": attr.label(default = "@io_bazel_rules_go//:stdlib"),
        "_go_context_data": attr.label(
            default = "@io_bazel_rules_go//:go_context_data",
        ),
        "_cgo_context_data": attr.label(default = "@io_bazel_rules_go//:cgo_context_data_proxy"),
        "_main_file_generator_tool": attr.label(
            executable = True,
            cfg = "exec",
            allow_files = True,
            default = Label("//go/pkg/beam/runners/vet/bazel/cmd/beambazel:beambazel"),
        ),
    },
    toolchains = ["@io_bazel_rules_go//go:toolchain"],
    doc = """This builds a Go library from a set of source files that are all part of
    the same package.<br><br>
    ***Note:*** For targets generated by Gazelle, `name` is typically the last component of the path,
    or `go_default_library`, with the old naming convention.<br><br>
    **Providers:**
    <ul>
      <li>[GoLibrary]</li>
      <li>[GoSource]</li>
      <li>[GoArchive]</li>
    </ul>
    """,
)
