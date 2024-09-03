# goRegex

A library that implements a Non Finite Automata based regex engine in Go Lang


### how to use: 
```go

pattern, err := rgx.Compile(regexString)
if err != nil {
	// error handling
}
results := pattern.FindMatches(content)

if results.Matches {
	groupMatchString := results.Groups["group-name"]
}
```

### Regex 

- [x] `^` beginning of the string
- [x] `$` end of the string
- [x] `.` any single character/wildcard
- [x] bracket notation
  - [x] `[ ]` bracket notation/ranges
  - [x] `[^ ]` bracket negation notation
  - [x] better handling of the bracket expressions: e.g., `[ab-exy12]`
  - [x] special characters in the bracket
    - [x] support escape character
- [x] quantifiers
  - [x] `*` none or more times
  - [x] `+` one or more times
  - [x] `?` optional
  - [x] `{m,n}` more than or equal to `m` and less than equal to `n` times
- [x] capturing group
  - [x] `( )` capturing group or subexpression
  - [x] `\n` backreference, e.g, `(dog)\1` where `n` is in `[0, 9]`
  - [x] `\k<name>` named backreference, e.g, `(?<animal>dog)\k<animal>`
  - [x] extracting the string that matches with the regex
- [x] `\` escape character


## references
- [How to build regex](https://rhaeguard.github.io/posts/regex/)
