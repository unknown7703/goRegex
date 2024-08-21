package goregex

func getChar(input string,pos int) uint8{
	if pos>=0 && pos < len(input){
		return input[pos]
	}
	if pos>=len(input){
		return endOfText
	}

	return startOfText
}

func (s *State) nextStartWith(ch uint8)* State{
	states := s.transitions[ch]
	if len(states) == 0 {
		return nil
	}
	return states[0]
}


