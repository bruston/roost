roost
=====

Roost is an **experimental** stack-based language similar to Forth.

## Usage

```
"Hello, World!" .
```

Outputs: `Hello, World!`

Pushes the string "Hello, World!" on the stack then pops and prints it to stdout.

## Arithmetic

Each operator expects two values on the stack. They are popped and replaced with the result of the operation.

```forth
5 10 + .
```

Outputs: `15`

```forth
10 5 - .
```

Outputs: `5`

```forth
10 5 * .
```

Outputs: `50`

```forth
6 3 / .
```

Outputs: `2`

```forth
6 2 % .
```

Outputs: `0`

## Loops

`for` expects two number values on the stack: limit and starting index. The loop body executes while limit < index. The `end` keyword marks the end of a loop body. Index is incremented each iteration. The current index value can be pushed on the stack using the special `I` word.

```forth
10 0 for I . LF . end
```

*Note*: The word `LF` pushes a newline character.

Outputs:

```
0
1
2
3
4
5
6
7
8
9
```
## Defining A Word

Word definition starts with a colon followed by a name and ends with a semicolon. Everything between the name and semicolon is considered the word body. A word body is comprised of other words and values.

```forth
: square dup * ;
```

*Note*: The word `dup` duplicates the value at the top of the stack.

## Conditionals

The keyword `if` pops a value, if it is the boolean value true the if body is executed before proceeding to the code following the `then` keyword.

```forth
: is-even? dup 2 % 0 = ;

6 is-even? if "foo" . then "bar" .
```

Outputs: `foobar`

An if statement may contain an optional else:

```forth
5 5 = if "five" else "not five" then .
```

Outputs: `five`

*Note*: The word `=` pops two values and compares them, pushing true if they are equal, else false.

## Variables

Most things are accomplished by manipulating the stack, but it is also possible to declare variables. Variables have global scope (at least for now).

A variable is declared using the `var` keyword.

```forth
var foo
```

This creates a new empty value with the name `foo` and pushes a reference to it on the stack. It also defines a new word with the same name. When executed, the word simply pushes a reference to the variable on the stack.  A reference is used with the `!` and `@` words.

With a reference to `foo` now on the stack, storing a variable looks like this:

```forth
"bar" !
```

`!` (store) expects to pop two values. First, "foo" the value to be stored then the variable reference. The value is stored in the `foo` variable declared previously.

Loading a value stored in a variable is done using the `@` (fetch) word.

```forth
foo @
```

The value stored in `foo` is pushed on the stack.

## Types

Roost supports the following types:

**String**

`"string"`

**Boolean**

`true`, `false`

**Number**

64-bit float.

`-5.5`, `5`, `-5`

**Byte**

Unsigned 8-bit int.

`'5'`, `'a'`

**Slice**

A collection of arbitrary values that grows as needed.

`{ 1 2 3 "foo" "bar" false }`

New values are inserted at the end.

## Using Roost With Go

It is possible to embed roost in Go programs.

```go
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bruston/roost/parser"
	"github.com/bruston/roost/runtime"
)

func main() {
	listen := flag.String("listen", ":8080", "host:port to listen on")
	script := flag.String("script", "demo.roost", "path to roost script")
	flag.Parse()
	http.Handle("/hello/", scriptRunner{path: *script})
	log.Fatal(http.ListenAndServe(*listen, nil))
}

type scriptRunner struct {
	path string
}

func (s scriptRunner) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(s.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
	p := parser.New(f)
	ast, err := p.Parse()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	env := runtime.New(10)
	env.Stack.PushString(path.Base(r.URL.Path))
	if err := runtime.Eval(env, ast); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if env.Stack.Len() == 1 && env.Stack.Peek().Type() == runtime.ValueString {
		fmt.Fprintf(w, "%s\n", template.HTMLEscapeString(env.Stack.Pop().Value().(string)))
	}
}
```

With the script:

```forth
"Hello " swap +
```
