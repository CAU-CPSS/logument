# **_TSON_**

**_TSON_**, which stands for Time-Stamped JSON, is used to store timestamps as well as data values.

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

## VSCode Extension

There is a Visual Studio Code extension, which enables **_TSON_** syntax highlighting.  
Please refer to [VSCode-TSON-Extension](https://github.com/CAU-CPSS/VSCode-TSON-Extension) repository.
