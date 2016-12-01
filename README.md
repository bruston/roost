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
