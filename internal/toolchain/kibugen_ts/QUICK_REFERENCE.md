# Quick Reference: Making Go Slices Optional in TypeScript

## TL;DR

**Problem:** Go slices can be `nil`, but TypeScript arrays are generated as required fields.

**Solution:** Add one line of code to make slices optional.

## The One-Line Fix

**File:** `/Users/jqualls/projects/github.com/kibu-sh/kibu/internal/toolchain/kibugen_ts/kibugen_ts_types.go`

**Line:** ~140 (in the `typeFromSlice` function)

```go
func typeFromSlice(params *typeBuilderParams) (tsType string, isOptional bool, err error) {
	slice, ok := params.ty.(*types.Slice)
	if !ok {
		return
	}

	var elemType string
	elemType, _, err = params.dive(&typeBuilderParams{
		ctx:  params.ctx,
		ty:   slice.Elem(),
		dive: params.dive,
	})
	if err != nil {
		return
	}

	tsType = elemType + "[]"
	isOptional = true  // ← ADD THIS LINE
	return
}
```

## Before & After

### Go Input
```go
type CheckResponse struct {
    StatusList []Status `json:"status_list"`
}
```

### TypeScript Output

**Before (Current):**
```typescript
type CheckResponse = {
  status_list: Status[]  // Required - error if null
}
```

**After (Proposed):**
```typescript
type CheckResponse = {
  status_list?: Status[] | null  // Optional - handles nil slices
}
```

## Why This Matters

### Go Behavior
- `nil` slices JSON-marshal to `null`
- Empty slices JSON-marshal to `[]`
- Both are valid return values

### TypeScript Impact
- **Current:** Type error when Go returns `null`
- **After:** Correctly handles both `null` and `[]`

## Test It

```bash
cd /Users/jqualls/projects/github.com/kibu-sh/kibu/internal/toolchain/kibugen_ts
go test -v -run TestAnalyzer
```

Look for this change in the output:
```diff
- status_list: internal_toolchain_kibugen_ts_testdata_analyzer_health_Status[]
+ status_list?: internal_toolchain_kibugen_ts_testdata_analyzer_health_Status[] | null
```

## Bonus: Make Maps Optional Too

Maps in Go can also be `nil`. Apply the same fix:

**File:** Same file (`kibugen_ts_types.go`)

**Function:** `typeFromMap` (line ~192)

```go
tsType = fmt.Sprintf("Record<%s, %s>", keyType, valueType)
isOptional = true  // ← ADD THIS LINE
return
```

## Type Optionality Summary

| Go Type | Current TS | Should Be | Nil-able? |
|---------|-----------|-----------|-----------|
| `*T` | `T \| null` (optional) | ✅ Correct | Yes |
| `[]T` | `T[]` (required) | ❌ Should be optional | Yes |
| `[N]T` | `T[]` (required) | ✅ Correct | No |
| `map[K]V` | `Record<K,V>` (required) | ❌ Should be optional | Yes |

## Related Files

- **Main logic:** `kibugen_ts_types.go` (type conversion handlers)
- **Field generation:** `kibugen_ts.go` (uses isOptional flag)
- **Test data:** `testdata/analyzer/health/health.go`
- **Test:** `kibugen_ts_test.go`

## Architecture

The codebase uses a chain-of-responsibility pattern:
1. Each type handler (`typeFromSlice`, `typeFromPointer`, etc.) returns `(tsType, isOptional, err)`
2. The `isOptional` flag is automatically handled by field generation logic
3. Adding `isOptional = true` makes fields get `?` and `| null`

No other changes needed - the infrastructure is already there!
