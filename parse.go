package goregex

import (
	"fmt"
	"go/token"
	"strconv"
	"strings"
)

type rgTokenType uint8

const(
	literal         rgTokenType = iota // any literal character, e.g., a, b, 1, 2, etc.
	or                             = iota // |
	bracket                        = iota // []
	bracketNot                     = iota // [^]
	groupCaptured                  = iota // ()
	groupUncaptured                = iota // logical group
	wildcard                       = iota // .
	textBeginning                  = iota // ^
	textEnd                        = iota // $
	backReference                  = iota // $
	quantifier                     = iota // {m,n} or {m,}, {m}
)
// value can be anything 
type rgToken struct{
	tokenType rgTokenType
	value interface{}
}
// min and max will show the range of quant payload 
type quantPayload struct{
	min int
	max int 
	value rgToken
}

type groupPayload struct{
	token []rgToken
	name string
}
// store position and tokens also stored captured group and encountered group count
type parsingContext struct{
	pos int 
	tokens []rgToken
	groupCount uint8
	capturedGroups map[string]bool
}

// methods of parsing context 

// location
func (p* parsingContext)loc() int{
	return p.pos
}
// updates groupcount and return gc
func (p* parsingContext) nextGroup() uint8{
	p.groupCount++
	return p.groupCount
}
// advance to next position , iterator
func (p* parsingContext) adv() int {
	p.pos+=1
	return p.pos
}
// advance to specific position
func (p* parsingContext) advTo(pos int){
	p.pos=pos
}
// push new token into tokens[] of parsecontext
func (p* parsingContext) push(token rgToken){
	p.tokens = append(p.tokens, token)
}
// remove last 'n' tokens from tokens[] of parseContext
func (p* parsingContext) remLast(count int)[]rgToken {
	toRem :=p.tokens[len(p.tokens)-count:]
	p.tokens= append([]rgToken{},p.tokens[:len(p.tokens)-count]...)
	return toRem
}
/////////////////////////////////////
// checker functions
//aplha check
func isAlphaLow (ch uint8) bool{
	return ch>= 'a' && ch<='z'
}
//aplha check
func isAlphaUp (ch uint8) bool{
	return ch>= 'A' && ch<='Z'
}
// num check 
func isDig (ch uint8) bool{
	return ch>='0'&&ch<='9'
}

// charecter maps to find chaecter that defines the regex 
// eg r'@gmail.com$ dollar should be kept in map so parser can differentiate
// from rest of the string ie. @gmail.com
var specialChar =map[uint8]bool{
	'&':  true,
	'*':  true,
	' ':  true,
	'{':  true,
	'}':  true,
	'[':  true,
	']':  true,
	'(':  true,
	')':  true,
	',':  true,
	'=':  true,
	'-':  true,
	'.':  true,
	'+':  true,
	';':  true,
	'\\': true,
	'/':  true,
}
// regex relevant char
var mustBeEscapedChar = map[uint8]bool{
	'[':  true,
	'\\': true,
	'^':  true,
	'$':  true,
	'.':  true,
	'|':  true,
	'?':  true,
	'*':  true,
	'+':  true,
	'(':  true,
	')':  true,
	'{':  true,
	'}':  true,
}
// is in special char map or not
func isSpecial(ch uint8)bool {
	_, flag := specialChar[ch]
	return flag
}
// check all condition aplha num special char
func isLiteral(ch uint8)bool{
	 return isAlphaLow(ch)||isAlphaUp(ch)||isDig(ch)||isSpecial(ch)
}
// check for '.'
func isWild(ch uint8)bool{
	return ch=='.'
}

const quantInfinity =-1

// map quant symbols as {}range eg + is 1 to inf , * is 0 to inf
var quantToCurly= map[uint8][]int{
	'*':{0,quantInfinity},
	'+':{1,quantInfinity},
	'?':{0,1},
}
// is present as quant symbol - *,+,?
func isQuantifier(ch uint8)bool{
	_,ok:=quantToCurly[ch]
	return ok
}

//parce [] brackets , parses for inside content of []
func parseBracket(regString string,parCtx *parsingContext) *RegexError{
	var tokenType rgTokenType

	if regString[parCtx.pos]=='^'{
		tokenType= bracketNot
		parCtx.adv()
	}else{
		tokenType=bracket
	}
	//literals within []
	var pieces[]string
	for parCtx.loc()< len(regString) && regString[parCtx.loc()]!=']'{
		ch:= regString[parCtx.loc()]
		if ch=='-' && parCtx.loc()+1<len(regString){
			nextChar:= regString[parCtx.loc()+1]
			if(len(pieces)==0 || nextChar==']'){
				pieces = append(pieces, fmt.Sprintf("%c",ch))
			}else{
				parCtx.adv()
				piece :=pieces[len(pieces)-1]
				if(len(piece)==1){
					prevChar:=piece[0]
					if(prevChar<nextChar){
						pieces[len(pieces)-1]= fmt.Sprintf("%c%c",prevChar,nextChar)
					}else{
						return &RegexError{
							Code: SyntaxError,
							Message: fmt.Sprintf("'%c-%c' Range Not Acceptable", prevChar, nextChar),
							Pos: parCtx.loc(),
						}
					}
				}else{
					pieces = append(pieces, fmt.Sprintf("%c",ch))
				}
			}
		}else if ch=='\\' && parCtx.loc()+1<len(regString){
			nextChar:= regString[parCtx.adv()]
			pieces=append(pieces, fmt.Sprintf("%c",nextChar))
		}else{
			pieces=append(pieces, fmt.Sprintf("%c",ch))
		}
		parCtx.adv()
	}
	uniqueCharPieces:=map[uint8]bool{}
	for _,piece :=range pieces{
		for s:= piece[0];s<=piece[len(piece)-1];s++{
			uniqueCharPieces[s]=true
		}
	}
	token:=rgToken{
		tokenType: tokenType,
		value: uniqueCharPieces,
	}
	parCtx.tokens=append(parCtx.tokens, token)

	return nil
}
//////////////////////////////////////////////

//parse groups ()
func parseGroup(regString string,parCtx *parsingContext) *RegexError{
	groupContext:=parsingContext{
		pos: parCtx.loc(),
		tokens: []rgToken{},
	}
	
}


