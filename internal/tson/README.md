# TSON

TSON, which stands for Time-Stamped JSON, is used to store timestamps as well as data values.

## BNF of TSON

```bnf
<tson> ::= <object> | <array> | <value>

<object> ::= "{" <members>? "}"
<members> ::= <pair> ("," <pair>)*
<pair> ::= <string> <timestamp>? ":" <value>

<array> ::= "[" <elements>? "]"
<elements> ::= <value> ("," <value>)*

<value> ::= <primitive> <timestamp>? | <object> | <array> | "null"
<primitive> ::= <string> | <number> | <boolean>

<timestamp> ::= "<" <timestamp_value> ">"
```