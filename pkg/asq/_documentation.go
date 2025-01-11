// package asq
//
// # Wildcard comment tags
//
// A wildcard comment tag /***/ has an "active interval" starting at its start
// position, and ending at the start of the next wildcard comment tag, or the
// end of the query region.
//
// In the example below, there are 4 active intervals on this line of code.
//
// |-------|------------|--------------|----------------|----------|
// asq_start
// /***/if /***/x >= 10 /***/&& y < 20 /***/{ foo.Bar()./***/Baz() }
// asq_end
//
// The /***/ wildcard tag will apply to the 1st syntactic entity inside its
// active interval. So in the example above, the 1st wildcard tag applies to
// the entire if statement. The 2nd wildcard tag applies to the variable x.
// The 3rd wildcard tag applies to the binary operator &&. The 4th applies to
// the entire code block of the if statement. The 5th wildcard tag applies to
// the call to the method Baz.
package asq
