package main

import (
	"bytes"
	"fmt"
)

const (
	Common = iota
	MayBeComment
	MayBeCommentOver
	Comment
)

func removeComments(sources []string) []string {
	state := Common
	newSource := []string{}
	lineSource := bytes.NewBuffer([]byte{})
	for _, source := range sources {
		if state != Comment {
			if lineSource.Len() >0 {
				newSource = append(newSource, lineSource.String())
				lineSource.Reset()
			}
		}

		for i:=0;i< len(source);i++ {
			if source[i] == '/' {  //may be comment start may be comment  / /*  */
				switch state {
				case Common:
					state = MayBeComment   //Common - > MayBeComment
					lineSource.WriteByte(source[i])
				case MayBeComment:
					lineSource.Truncate(lineSource.Len()-1)
					state = Common // MayBeComment -> Comment - > Common
					goto NEWLINE   // skip line
				case Comment:
					continue
				case MayBeCommentOver:
					state = Common // MayBeComment -> Comment - > Common
					continue       //skip char
				}
			} else if source[i] == '*' {   /*  */
				switch state {
				case Common:
					lineSource.WriteByte(source[i])
				case MayBeComment:
					lineSource.Truncate(lineSource.Len()-1)
					state = Comment
					continue //skip char
				case Comment:
					state = MayBeCommentOver
					continue //skip char
				case MayBeCommentOver:
					state = MayBeCommentOver
				}
			} else{
				switch state {
				case Common:
					lineSource.WriteByte(source[i])
				case MayBeComment:
					lineSource.WriteByte(source[i])
					state = Common
					continue //skip char
				case Comment:
					continue //skip char
				case MayBeCommentOver:
					state = Comment
				}
			}
		}
		NEWLINE:
	}
	if lineSource.Len() >0 {
		newSource = append(newSource, lineSource.String())
	}
	return newSource
}

func main() {
	source := []string{"void func(int k) {", "// this function does nothing /*", "   k = k*2/4;", "   k = k/2;*/", "}"}
	fmt.Println(removeComments(source))
}