---
title: Lokomotive coding style guide
weight: 10
---

This document describes the coding style that's enforced in the Lokomotive codebase. The guidelines
in this document should be followed by Lokomotive maintainers and contributors alike. The guide is
structured in a way which makes it easy to link to a specific guideline, for example from a PR
comment.

## Guidelines

This section describes the entire set of coding style guidelines which should be followed when
working on the Lokomotive codebase.

Each "topic" should reside on its own under a heading so that it's easy to unambiguously link to a
specific guideline during code reviews.

The list of guidelines is meant to evolve over time. If you think a new guideline should be
enforced, suggest changes to this document in a PR.

### Error string formatting

Error strings are typically *chained*:

```
Deploying cluster: reading component config: parsing HCL: unexpected character '%'
```

Remember this fact when constructing error strings and think about how the final string might look
to the user.

#### Wrapping errors

When "wrapping" an error in another error to provide context, the `%w` formatting verb should be
used following the [Go 1.13 recommendations](https://blog.golang.org/go1.13-errors) to error
handling:

`fmt.Errorf("parsing config: %w", err)`

#### Error string text

Don't include words which describe failure when *wrapping* errors. This includes words such as
"failed", "error", "could not" etc. It's enough to mention a failure in the **beginning** of the
error string, i.e. when logging the error, and if necessary also at the **end** of the error
string, i.e. when creating the error value at the deepest level of the call stack. There is no need
to repeat the fact that something failed in every intermediate "piece" of the error string.

To demonstrate this, consider the following:

```
Error deploying cluster: error reading component config: failed to parse HCL: unexpected character '%' while parsing HCL
```

vs.

```
Error deploying cluster: reading component config: parsing HCL: unexpected character '%'
```

The 2nd example provides the same information as the 1st one without repetition.

When formatting an error string, provide just the *context*, i.e. **what we were doing** when the
error occurred (or the erroneous condition if the error is the last one in the "chain").

Good:

- `fmt.Errorf("parsing config: %w", err)`
- `fmt.Errorf("deploying cluster: %w", err)`
- `fmt.Errorf("reversing time: %w", err)`
- `fmt.Errorf("space probe is offline")`

Bad:

- `fmt.Errorf("parsing config failed: %w", err)`
- `fmt.Errorf("error encountered while deploying cluster: %w", err)`
- `fmt.Errorf("could not reverse time: %w", err)`

#### Capitalizing error strings

Don't capitalize error strings when *returning* an error (remember - errors are chained).
Capitalize the **beginning** of an error string when **logging** an error.

Good:

- `fmt.Errorf("parsing config: %w", err)`
- `log.Errorf("Deploying cluster: %v", err)`

Bad:

- `fmt.Errorf("Parsing config: %w", err)`
- `log.Errorf("deploying cluster: %v", err)`

### Comments

Follow the standard Go [guidelines](https://blog.golang.org/godoc) for comments in Go code. *In
addition*, follow the guidelines below.

#### Whitespace

Include a space between `//` and the beginning of the comment.

Good:

```go
// This is a comment.
```

Bad:

```go
//This is a comment.
```

#### Capitalization

Capitalize the first letter of a Go comment unless the official Go guidelines require you to do
otherwise.

Good:

```go
// A very important constant.
const this = "that"
```

Bad:

```go
// a very important constant.
const this = "that"
```

#### Periods

As a general rule, end all comments with a period. Possible exceptions are *very* short comments
(up to 2-3 words which don't form a self-contained sentence) which are typically **inline**,
however generally you should format comments as full sentences ending with a period.

Good:

```go
// Foo does this and also that.
func Foo() {
    ...
    n := 3 // Default
}
```

Bad:

```go
// Foo does this and also that
func Foo() {
    ...
}
```

#### "nolint" comments

"nolint" comments are a special type of comment used to tell the linter to ignore a specific
linting rule. They are formatted differently from "regular" Go comments.

Format "nolint" comments according to the following example:

```go
//nolint:funlen,gosec
func foo() {
    ...
}
```

- **Don't** put a space between `//` and `nolint`.
- Don't capitalize "nolint" comments.
- Don't end "nolint" comments with a period.
- Don't put spaces after commas when multiple linting rules are ignored.

When adding a "nolint" comment for a function which has a doc comment, put the "nolint" line below
the doc comment after a "separator" comment line:

```go
// Foo does this and that.
//
//nolint:funlen,gosec
func Foo() {
    ...
}
```

### Table-driven tests

Whenever possible, use [table-driven tests](https://github.com/golang/go/wiki/TableDrivenTests).
Table-driven tests are compact, readable and easy to maintain compared to test cases that are
described imperatively.

It is usually possible to construct most tests in a way which allows the various test cases to be
described *declaratively* in a Go struct. Most tests can be broken down to "steps" similar to the
following:

1. Set some initial state.
1. Call the function we're testing with specific arguments.
1. Ensure the return value matches a desired result.

Each step in the list above can correspond to a field in the Go struct which describes the test
cases. There are exceptions to this, but in general try to group similar test cases in one test
function. You can then share the test execution logic for all test cases as in the example below.

Test cases which require a different *execution logic* (e.g. a different sequence of function
calls) should reside in separate test functions.

**Example**

```go
func add(a, b int) int {
	return a + b
}

func TestAdd(t *testing.T) {
	tests := []struct {
		desc string
		a    int
		b    int
		want int
	}{
		{
			desc: "Positive numbers",
			a:    1,
			b:    3,
			want: 4,
		},
		{
			desc: "Negative and positive numbers",
			a:    1,
			b:    -3,
			want: -2,
		},
		{
			desc: "Zero and a positive number",
			a:    0,
			b:    5,
			want: 5,
		},
		{
			desc: "Failing test",
			a:    2,
			b:    2,
			want: 5,
		},
	}

	for _, test := range tests {
		// Pin local var to avoid scope issues.
		test := test

		t.Run(test.desc, func(t *testing.T) {
			sum := add(test.a, test.b)

			if sum != test.want {
				t.Fatalf("Unexpected result: got %d, want %d", sum, test.want)
			}
		})
	}
}
```

Using a structure as in the example above provides an easy-to-read output:

```
=== RUN   TestAdd
=== RUN   TestAdd/Positive_numbers
=== RUN   TestAdd/Negative_and_positive_numbers
=== RUN   TestAdd/Zero_and_a_positive_number
=== RUN   TestAdd/Failing_test
    main_test.go:46: Unexpected result: got 4, want 5
--- FAIL: TestAdd (0.00s)
    --- PASS: TestAdd/Positive_numbers (0.00s)
    --- PASS: TestAdd/Negative_and_positive_numbers (0.00s)
    --- PASS: TestAdd/Zero_and_a_positive_number (0.00s)
    --- FAIL: TestAdd/Failing_test (0.00s)
FAIL
FAIL    _/tmp/go        0.001s
FAIL
```
