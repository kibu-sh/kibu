# kibumod

Static analyzer. kibugenv2 and kibugen_ts require it. Do not run kibumod by hand.

It scans packages for interfaces with `//kibu:service`, `//kibu:workflow`, or `//kibu:activity`, parses decorator options, and returns a `modspecv2.Package`.

If generation finds nothing, the interface is missing a type-level `//kibu:` decorator or lives in a package outside the `./...` pattern.

Decorators: [decorators.md](decorators.md).
