// Package readstruct is a simple tool to read (some of) struct info from source files.
//
// To keep codes short and simple, the struct *MUST* follow these rules:
//
//   1. For non-embed fields, type must be either identifier or unnamed struct.
//   2. For embed fields, type must be an identifier.
package readstruct
