package fit

var binaryOperators = map[string]bool{
	"+": true, "-": true, "*": true, "/": true, "%": true,
	"<": true, "<=": true, ">": true, ">=": true, "==": true, "!=": true,
	"&&": true, "||": true,
}

var unaryOperators = map[string]bool{"!": true, "-": true}
