package goregex

import (
	"fmt"
)

type group struct{
	names []string
	start   bool
	end     bool
}

type backreference struct{
	name string
	target *State
}

type State struct{
	start         bool
	terminal      bool
	endOfText     bool
	startOfText   bool
	transitions   map[uint8][]*State
	groups        []*group
	backreference *backreference
}

const (
	epsilonChar = 0
	startOfText = 1
	endOfText   = 2
	anyChar     = 3
	newline     = 10
)
////////////////////////////
func tokenToNfa(token rgToken,parCtx *parsingContext,startFrom *State)(*State ,*State,*RegexError){
	switch token.tokenType{
	case literal:
		value:=token.value.(uint8)
		to:=&State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[value]=[]*State{to}
		return startFrom,to,nil
	case quantifier:
		return handleQuantifier(token,parCtx,startFrom)
	case wildcard:
		to:= &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.transitions[anyChar]=[]*State{to}
		return startFrom,to,nil
	case or:
		value :=token.value.([]rgToken)
		_,end1,err:=tokenToNfa(value[0],parCtx,startFrom)
		if err!=nil{
			return nil,nil,err
		}
		_,end2,err:=tokenToNfa(value[1],parCtx,startFrom)
		if err!=nil{
			return nil,nil,err
		}
		to:=&State{
			transitions: map[uint8][]*State{},
		}
		end1.transitions[epsilonChar]=append(end1.transitions[epsilonChar], to)
		end2.transitions[epsilonChar]=append(end1.transitions[epsilonChar], to)
		return startFrom,to,nil
	case groupCaptured:
		v:=token.value.(groupPayload)
		start,end,err:=tokenToNfa(v.token[0],parCtx,&State{
			transitions: map[uint8][]*State{},
		})

		if err !=nil{
			return nil,nil,err
		}
		for i:=1;i<len(v.token);i++{
			_, endNext,err:=tokenToNfa(v.token[i],parCtx,end)
			if err!=nil{
				return nil, nil, err
			}
			end=endNext
		}
		groupNameNumeric:= fmt.Sprintf("%d",parCtx.nextGroup())
		groupNameUserSet := v.name
		groupNames := []string{groupNameNumeric}
		parCtx.capturedGroups[groupNameNumeric] = true
		if groupNameUserSet != "" {
			groupNames = append(groupNames, groupNameUserSet)
			parCtx.capturedGroups[groupNameUserSet] = true
		}

		if startFrom.groups != nil {
			startFrom.groups = append(startFrom.groups, &group{
				names: groupNames,
				start: true,
			})
		} else {
			startFrom.groups = []*group{{
				names: groupNames,
				start: true,
			}}
		}

		if end.groups != nil {
			end.groups = append(end.groups, &group{
				names: groupNames,
				end:   true,
			})
		} else {
			end.groups = []*group{{
				names: groupNames,
				end:   true,
			}}
		}

		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], start)
		return startFrom, end, nil
	case groupUncaptured:
		values:=token.value.([]rgToken)

		if len(values)==0{
			end:=&State{
				transitions: map[uint8][]*State{},
			}
			startFrom.transitions[epsilonChar]=append(startFrom.transitions[epsilonChar], end)
			return startFrom,end,nil
		}
		start ,end,err:= tokenToNfa(values[0],parCtx,&State{
			transitions: map[uint8][]*State{},
		})
		if err!=nil{
			return nil,nil,err
		}
		
		for i := 0; i < len(values); i++ {
			_,endNext,err :=tokenToNfa(values[i],parCtx,end)
			if err!=nil{
				return nil,nil,err
			}
			end=endNext
		}
		startFrom.transitions[epsilonChar]=append(startFrom.transitions[epsilonChar], start)
		return startFrom,end,nil
	case bracket:
		constructToken:= token.value.(map[uint8]bool)
		to:=&State{
			transitions: map[uint8][]*State{},
		}
		for ch:=range constructToken{
			startFrom.transitions[ch]=[]*State{to}
		}
		return 	startFrom,to,nil
	case bracketNot:
		constructTokens := token.value.(map[uint8]bool)

		to := &State{
			transitions: map[uint8][]*State{},
		}

		deadEnd := &State{
			transitions: map[uint8][]*State{},
		}

		for ch := range constructTokens {
			startFrom.transitions[ch] = []*State{deadEnd}
		}
		startFrom.transitions[anyChar] = []*State{to}

		return startFrom, to, nil
	case textBeginning:
		to := &State{
			transitions: map[uint8][]*State{},
		}
		startFrom.startOfText = true
		startFrom.transitions[epsilonChar] = append(startFrom.transitions[epsilonChar], to)
		return startFrom, to, nil
	case textEnd:
		startFrom.endOfText = true
		return startFrom, startFrom, nil
	case backReference:
		groupName := token.value.(string)
		if _, ok := parCtx.capturedGroups[groupName]; !ok {
			return nil, nil, &RegexError{
				Code:    CompilationError,
				Message: fmt.Sprintf("Group (%s) does not exist", groupName),
			}
		}
		to := &State{
			transitions: map[uint8][]*State{},
		}

		startFrom.backreference = &backreference{
			name:   groupName,
			target: to,
		}

		return startFrom, to, nil
	default:
		return nil, nil, &RegexError{
			Code:    CompilationError,
			Message: fmt.Sprintf("unrecognized token: %+v", token),
		}
	} 
}
func handleQuantifier(token rgToken,parCtx *parsingContext, startFrom *State)(*State,*State,*RegexError){
	payload:=token.value.(quantPayload)
	min:= payload.min
	max:= payload.max
	to:= &State{
		transitions: map[uint8][]*State{},
	}

	if min==0{
		startFrom.transitions[epsilonChar]= append(startFrom.transitions[epsilonChar], to)
	}

	var total int

	if max!=quantInfinity{
		total = max
	}else{
		if min == 0{
			total=1
		}else{
			total=min
		}
	}
	var value=payload.value
	previousStart,previousEnd,err:= tokenToNfa(value,parCtx, &State{
		transitions: map[uint8][]*State{},
	})

	if err!=nil{
		return nil,nil,err
	}
	startFrom.transitions[epsilonChar]=append(startFrom.transitions[epsilonChar], previousStart)

	for i:= 2;i<=total;i++{
		start,end,err := tokenToNfa(value,parCtx,&State{
			transitions: map[uint8][]*State{},
		})
		if err!=nil{
			return nil,nil,err
		}

		previousEnd.transitions[epsilonChar]=append(previousEnd.transitions[epsilonChar], start)

		previousStart=start
		previousEnd=end

		if i>min{
			start.transitions[epsilonChar]=append(start.transitions[epsilonChar], to)
		}

	}
	previousEnd.transitions[epsilonChar]=append(previousEnd.transitions[epsilonChar], to)
	if max == quantInfinity{
		to.transitions[epsilonChar] = append(to.transitions[epsilonChar], previousStart)
	}
	return startFrom,to,nil
}
////////////////////////////
func toNfa(parCtx *parsingContext)(*State,*RegexError){
	token := parCtx.tokens[0]

	startState,endState,err := tokenToNfa(token,parCtx,&State{
		transitions: map[uint8][]*State{},
	})
	if err!=nil{
		return nil,err
	}
	for i:=1; i<len(parCtx.tokens);i++{
		_,endNext,err:= tokenToNfa(parCtx.tokens[i],parCtx,endState)
		if err!=nil{
			return nil,err
		}
		endState=endNext
	}
	start :=&State{
		start: true,
		transitions: map[uint8][]*State{
			epsilonChar: {startState},
		},
		groups: []*group{{
			names: []string{"0"},
			start: true,
			end:   false,
		}},
	}

	end := &State{
		transitions: map[uint8][]*State{},
		terminal: true,
		groups: []*group{
			{
				names: []string{"0"},
			start: false,
			end:   true,
			},
		},
	}

	endState.transitions[epsilonChar]=append(endState.transitions[epsilonChar], end)
	return start ,nil
}

