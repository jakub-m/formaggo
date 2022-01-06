# 🧀 Formaggo. A cheesy exhaustive state checker in Go.

Formaggo is a simple model changes inspired by TLA+. The checker and the models
are written in Go, compiled and run as a single binary.

A model is specified as a set of Transitions. A transition is a function that
takes a State on input and that returns possible states on output.  The checker
then builds a graph of all the possible states and reports violations of
Invariants and Properties.

The State is an arbitrary primitive or structure (an `interface{}`) that can be
hashed. Important! Private fields (the ones starting with the lower case) are
IGNORED when calculating hash.

State hash is calculated using [mitchellh/hashstructure][ref_hash]. Tags from
there apply, in particular:

```
* "set" - The field will be treated as a set, where ordering doesn't
          affect the hash code. This only works for slices.
```

E.g.:

```
struct {
    values []string `hash:"set"`
}
```


[ref_hash]:https://pkg.go.dev/github.com/mitchellh/hashstructure/v2


# FAQ

Q: RATS! What's that?

It is a bug. The code should never reach RATS path.

