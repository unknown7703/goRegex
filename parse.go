package goregex

import (
	"fmt"
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
	groupName:=""
	if regString[groupContext.loc()]=='?'{
		if regString[groupContext.adv()]=='<'{
			for regString[groupContext.adv()]!='>'{
				ch:=regString[groupContext.loc()]
				groupName+=fmt.Sprintf("%c",ch)
			}
		}else{
			return &RegexError{
				Code: SyntaxError,
				Message: "group name invalid",
				Pos: groupContext.loc(),
			}
		}
		groupContext.adv()
	}

	for groupContext.loc()<len(regString) && regString[groupContext.loc()]!=')'{
		ch:=regString[groupContext.loc()]
		if err:=processChar(regString,&groupContext,ch); err!=nil{
			return err
		}
		groupContext.adv()
	}
	if regString[groupContext.loc()] != ')' {
		return &RegexError{
			Code:    SyntaxError,
			Message: "Group has not been properly closed",
			Pos:     groupContext.loc(),
		}
	}

	token := rgToken{
		tokenType: groupCaptured,
		value: groupPayload{
			token: groupContext.tokens,
			name:   groupName,
		},
	}
	parCtx.push(token)
	parCtx.advTo(groupContext.loc())
	return nil
}
/////////////////////////////////////
//parse quantifiers
func parseQuant(ch uint8,parCtx *parsingContext){
	bound :=quantToCurly[ch]
	token :=rgToken{
		tokenType: quantifier,
		value: quantPayload{
			min: bound[0],
			max: bound[1],
			value: parCtx.remLast(1)[0],
		},
	}
	parCtx.push(token)
}
////////////////////////////////////
//parse quants {}
func parseBounded(rgString string,parCtx *parsingContext) *RegexError{
	starPos:=parCtx.adv()
	endPos:=parCtx.loc()
	for rgString[endPos]!='}'{
		endPos++
	}
	parCtx.advTo(endPos)
	rang :=rgString[starPos:endPos]
	pieces :=strings.Split(rang,",")
	if len(pieces)==0{
		return &RegexError{
			Code: SyntaxError,
			Message: "Atleast one bound required",
			Pos: starPos,
		}
		
	}
	var start int
	var end int
	var err error
	if len(pieces)==1{
		start,err = strconv.Atoi(pieces[0])
		if err!=nil{
			return &RegexError{
				Code: SyntaxError,
				Message: err.Error(),
				Pos: starPos,
			}
		}
		end=start
	}else if len(pieces)==2{
		start,err = strconv.Atoi(pieces[0])
		if err!=nil{
			return &RegexError{
				Code: SyntaxError,
				Message: err.Error(),
				Pos: starPos,
			}
		}
		if(pieces[1]==""){
			end = quantInfinity
		}else{
			end,err = strconv.Atoi(pieces[1])
			if err!=nil{
				return &RegexError{
					Code: SyntaxError,
					Message: err.Error(),
					Pos: starPos,
				}
			}
		}
	}
	token :=rgToken{
		tokenType: quantifier,
		value: quantPayload{
			min: start,
			max: end,
			value: parCtx.remLast(1)[0],
		},
	}
	parCtx.push(token)
	return nil
}
////////////////////////////////////////////
//parse backslash

func parseBackslash(regString string,parCtx *parsingContext) * RegexError{
	nextChar := regString[parCtx.loc()+1]
	if isDig(nextChar) { // cares about the next single digit
		token := rgToken{
			tokenType: backReference,
			value:     fmt.Sprintf("%c", nextChar),
		}
		parCtx.push(token)
		parCtx.adv()
	} else if nextChar == 'k' { // \k<name> reference
		parCtx.adv()
		if regString[parCtx.adv()] == '<' {
			groupName := ""
			for regString[parCtx.adv()] != '>' {
				nextChar = regString[parCtx.loc()]
				groupName += fmt.Sprintf("%c", nextChar)
			}
			token := rgToken{
				tokenType: backReference,
				value:     groupName,
			}
			parCtx.push(token)
			parCtx.adv()
		} else {
			return &RegexError{
				Code:    SyntaxError,
				Message: "Invalid backreference syntax",
				Pos:     parCtx.loc(),
			}
		}
	} else if _, canBeEscaped := mustBeEscapedChar[nextChar]; canBeEscaped {
		token := rgToken{
			tokenType: literal,
			value:     nextChar,
		}
		parCtx.push(token)
		parCtx.adv()
	} else {
		if nextChar == 'n' {
			nextChar = '\n'
		} else if nextChar == 't' {
			nextChar = '\t'
		}
		token := rgToken{
			tokenType: literal,
			value:     nextChar,
		}
		parCtx.push(token)
		parCtx.adv()
	}

	return nil
}
/////////////////////////////////////////////
//parse literal
func parseLiteral(ch uint8,parCtx *parsingContext){
	token := rgToken{
		tokenType: literal,
		value: ch,
	}
	parCtx.push(token)
}
/////////////////////////////////////////////
//parse group uncaptured
func parseGroupUncaptured(regString string,parCtx *parsingContext)* RegexError{
	groupCtx := parsingContext{
		pos: parCtx.loc(),
		tokens: []rgToken{},
	}
	for groupCtx.loc()<len(regString) && regString[groupCtx.loc()]!=')'{
		ch:=regString[groupCtx.loc()]
		if err:=processChar(regString,&groupCtx,ch);err!=nil{
			return err
		}
		groupCtx.adv()
	}
	token :=rgToken{
		tokenType: groupUncaptured,
		value: groupCtx.tokens,
	}
	parCtx.push(token)

	if groupCtx.loc() >= len(regString){
		parCtx.advTo(groupCtx.loc())
	}else if regString[groupCtx.loc()]==')'{
		parCtx.advTo(groupCtx.loc()-1)
	}
	return nil
}
// process all incoming char
func processChar(regString string,parCtx *parsingContext,ch uint8) *RegexError{
	if ch=='('{
		parCtx.adv()
		if err:=parseGroup(regString,parCtx); err!=nil{
			return nil
		}
	}else if ch=='['{
		parCtx.adv()
		if err:=parseBracket(regString,parCtx);err!=nil{
			return nil
		}
	}else if isQuantifier(ch){
		parseQuant(ch,parCtx)
	}else if ch=='{'{
		if err:=parseBounded(regString,parCtx);err!=nil{
			return nil
		}
	}else if ch=='\\'{
		if err:=parseBackslash(regString,parCtx);err!=nil{
			return nil
		}
	}else if isWild(ch){
		token:=rgToken{
			tokenType: wildcard,
			value: ch,
		}
		parCtx.push(token)
	}else if isLiteral(ch){
		parseLiteral(ch,parCtx)
	}else if ch=='|'{
		//left side of OR
		left:=rgToken{
			tokenType: groupUncaptured,
			value: parCtx.remLast(len(parCtx.tokens)),
		}
		//OR itself
		parCtx.adv()
		if err:= parseGroupUncaptured(regString,parCtx); err!=nil{
			return err
		}
		//right side of OR
		right:=parCtx.remLast(1)[0]

		token :=rgToken{
			tokenType: or,
			value: []rgToken{left,right},
		}
		parCtx.push(token)
	}else if(ch=='^'){
		token :=rgToken{
			tokenType: rgTokenType(textBeginning),
			value: ch,
		}
		parCtx.push(token)
	}else if(ch=='&'){
		token :=rgToken{
			tokenType: rgTokenType(textEnd),
			value: ch,
		}
		parCtx.push(token)
	}
	return nil
}

// parsing the content for finding the regex string
func parse(regString string, parCtx *parsingContext) *RegexError {
	for parCtx.loc() < len(regString) {
		ch := regString[parCtx.loc()]
		if err := processChar(regString, parCtx, ch); err != nil {
			return err
		}
		parCtx.adv()
	}
	return nil
}

