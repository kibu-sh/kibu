# Typed enum declaration experiment

`kibuenum.Analyzer` extracts enum declarations into `*kibuenum.Result` using
`analysis.Pass.TypesInfo`. It requires `inspect.Analyzer` and exposes `FromPass`,
matching the existing toolchain's analyzer conventions. It never executes source
or imports a project into the analyzer process. Tests load and type-check real Go
source, including functions that would panic if executed.

The isolated constructor is `github.com/kibu-sh/kibu/pkg/enum/experimental.Define`.
The existing `pkg/enum` API and all production generators are unchanged.

## Example and extracted data

The complete, compilable example is embedded in [the extraction script](testdata/scripts/000000_extract_enums.txtar).
It includes string and uint64 declarations, labels, descriptions, constant
references, inherited `iota` expressions, and hand-written receiver delegation.

```go
import enum "github.com/kibu-sh/kibu/pkg/enum/experimental"

type OrderStatus string
const OrderStatusPending OrderStatus = "pending"

var OrderStatuses = enum.Define[OrderStatus]([]enum.Member[OrderStatus]{
    {Value: OrderStatusPending, Label: "Awaiting payment"},
})

func (s OrderStatus) Validate() error { return OrderStatuses.Validate(s) }
```

For the complete fixture, the result contains the following data (source spans
and repeated package paths omitted here for readability):

```text
Declaration: OrderStatuses; Type: OrderStatus; Underlying: string
  OrderStatusPending => "pending"; Label: "Awaiting payment"
  OrderStatusPaid    => "paid";    Label: "Payment received"
  OrderStatusFailed  => "failed";  Label: "Payment failed"
    Description: "The payment attempt was unsuccessful."

Declaration: EventCodes; Type: EventCode; Underlying: uint64
  EventDeleted => "9223372036854775810"; Label: "Deleted"
  EventCreated => "9223372036854775808"; Label: "Created"
```

The omitted `OrderStatusInternal` and `EventUpdated` constants are not members.
`WireValue` is a decoded Go string for string enums and exact decimal text for
integer enums, including after JSON serialization. Consumers distinguish them
using `Underlying`; they must not automatically convert large integers to JS
numbers. Symbols include package path, name, and declaration position. Enum,
member, value, label, and description expressions include source spans. An
omitted metadata field has an empty string and zero source span.

## Accepted static grammar

- A direct call to the resolved SDK `Define` function, initializing one named
  package-level variable. Grouped `var` declarations are supported.
- One explicit generic argument: a local, defined, non-generic string or integer
  type. Aliases resolve to their canonical defined type.
- One literal slice of the SDK's instantiated `Member[T]`, with keyed member
  struct literals. Fully spelled and elided member types work. Slice aliases work; distinct
  defined slice types are excluded. Empty slices work.
- An explicit `Value` referring to a constant symbol of exactly type `T`.
  `go/types` supplies its evaluated value, including expressions, references and
  inherited `iota` declarations. Literal values and conversions at the member
  site are deliberately excluded so the constant symbol identity is preserved.
- Optional `Label` and `Description` fields that are compile-time string
  expressions, including constant references and concatenation.

Import aliases, dot imports and parentheses do not change resolved identity.
An unrelated local or imported function named `Define` is not a declaration.
Members remain in slice order; declarations are sorted by filename and source
offset, independent of `pass.Files` order.

Dynamic metadata, slice/member variables or function calls, positional member
fields, sparse indexed slices, local/nested calls, inferred type arguments,
primitive enum types, and foreign enum definitions produce categorized source
diagnostics. Duplicate wire values invalidate the declaration. Multiple valid
declarations for one type diagnose a conflict and exclude all declarations of
that type. Invalid declarations are omitted; consumers must honor diagnostics
before generation. Wrong typed constants and wrong generic member types normally
fail Go type checking first. `RunDespiteErrors` is intentionally disabled.

Function aliases and wrappers are not discovered: only calls resolving directly
to the SDK function count. The prototype does not follow later assignment to the
set variable or infer global membership from arbitrary runtime code. Runtime
`Define` itself accepts normal Go expressions; the analyzer imposes the narrower
declaration grammar. This experiment establishes initial declarations, not whole
program runtime equivalence.

## Runtime and integration

The experimental set provides `Has`, `Validate`, `Get`, ordered `Values` and
`Members`. Membership is immutable through its API; input and output slices are
copied. Invalid membership returns `*InvalidValueError[T]`; duplicate values panic
at construction with `*DuplicateValueError[T]` containing both indices and the
value. Receivers delegate to this same set, without a generated membership table.

The recommended next integration point is `kibumod`: move these target-neutral
model types into `modspecv2`, add `Package.Enums`, and have `kibumod.Analyzer`
require `kibuenum.Analyzer` and attach its result. That preserves a single shared
spec for Go and TypeScript consumers. This proof deliberately does not add that
dependency, change alias lowering, or generate TypeScript or Go receivers.

## Verification

Run within the project's Nix development shell (`go1.27.0`,
`GOTOOLCHAIN=local` for this verification):

```sh
go test ./pkg/enum/... ./internal/toolchain/kibuenum
```

The Nix configuration was present only in the primary checkout when this
worktree was created. Verification used `nix develop path:/path/to/primary/checkout
--command ...`, retaining this worktree as the current directory.

Analyzer tests follow the sibling toolchain testscript convention: six txtar
archives define real Go modules, invoke `kibuenum` through `pipeline.Run`, and use
`cmp`/`grep` to assert the extracted spec and diagnostics. The fixtures cover
symbols, exact values, metadata, locations, explicit membership, JSON precision,
supported and rejected grammar, and deterministic ordering with reversed input files and a rebuilt inspector.
Diagnostic goldens compare exact messages and both source-span endpoints;
regressions cover slice aliases and three members sharing one value. The SDK is resolved
from this checkout through a module replacement; no synthetic importer is used.
Ordinary runtime-library unit tests check membership, structured errors and
defensive copying.
