// Package sqlt defines an extension over text/templates for producing safe SQL.
//
// Like package `html/template`, this package parses normal text templates
// and then modifies the syntax tree to alter potentially unsafe components.
//
// By default, any value interpolation gets a special internal formatter
// function applied to it that map the value to a named parameter to
// prevent SQL injection vulnerabilities.
package sqlt
