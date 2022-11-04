#!/usr/bin/env node
// eslint no-var: 0
// jshint esversion: 6
"use strict";

import fs from "fs";
import parser from "@babel/parser";

// See https://github.com/babel/babel/issues/13855
import _traverse from "@babel/traverse";
const traverse = _traverse.default;


function locationString(node) {
    return (node.loc !== null) ? `[${node.loc.start.line},${node.loc.start.column}]` : "[]";
}

const parseOutputLines = [];

function logJSON(type, subtype, name, pos, array, extra = null) {
    const extraValue = (extra !== null) ? `${JSON.stringify(extra)}` : "{}";
    const arrayValue = (array !== null) ? array : false;
    const json = `{"type":"${type}","subtype":"${subtype}","data":${JSON.stringify(name)},"pos":${pos},` +
        `"array":${arrayValue}, "extra":${extraValue}}`;
    parseOutputLines.push(json);
}

function logIdentifierJSON(subtype, name, pos, extra = null) {
    logJSON("Identifier", subtype, name, pos, null, extra);
}

function logLiteralJSON(subtype, value, pos, inArray, extra = null) {
    logJSON("Literal", subtype, value, pos, inArray, extra);
}

function logParameter(p) {
    if (p === null) {
        return;
    }
    if (p.type === "Identifier") {
        // simple function parameter name: function f(x)
        logIdentifierJSON("Parameter", p.name, locationString(p));
    } else if (p.type === "AssignmentPattern" && p.left.type === "Identifier") {
        // parameter with default value: function f(x = 3)
        logIdentifierJSON("Parameter", p.left.name, locationString(p));
    }
}

function logParameters(node) {
    for (let p of node.params) {
        logParameter(p);
    }
}

function traverseAst(ast, printDebug = false) {
    /*
      The way this function is currently structured, the visitor finds identifiers by visiting every different type
      of relevant symbol, and then logging the relevant identifier node (e.g. id, label) for that symbol.
      This structure was kept from the previous switch statement version.

      However, a simpler way to do it would be to just traverse until an actual Identifier node is found,
      then figure what kind of identifier it is by inverting the logic for each case that currently is here.
      This should be possible since babel-traverse lets you access parent nodes,
      and it may simplify the code a little bit.

      TODO
       1. If this code needs to be extended much more, we should switch to the second way above.
       2. Consider adding state to allow distinction between elements from different arrays
       3. Consider logging names of decorators
     */

    const arrayVisitor = {
        StringLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteralJSON("String", path.node.value, loc, true, path.node.extra);
        },
        NumericLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteralJSON("Numeric", path.node.value, loc, true, path.node.extra);
        },
        TemplateElement: function(path) {
            const loc = locationString(path.node);
            logLiteralJSON("StringTemplate", path.node.value.raw, loc, true, path.node.value);
        }
    };

    const astVisitor = {
        FunctionDeclaration: function (path) {
            if (path.node.id !== null) {
                logIdentifierJSON("Function", path.node.id.name, locationString(path.node.id));
            }
            logParameters(path.node);
        },
        AssignmentExpression: function (path) {
            if (path.node.left.type === "Identifier") {
                logIdentifierJSON("AssignmentTarget", path.node.left.name, locationString(path.node.left));
            }
        },
        ClassExpression: function (path) {
            if (path.node.id !== null) {
                logIdentifierJSON("Class", path.node.id.name, locationString(path.node.id));
            }
        },
        ClassMethod: function (path) {
            if (path.node.kind !== "Constructor" && path.node.key.type === "Identifier") {
                logIdentifierJSON("Method", path.node.key.name, locationString(path.node.key));
            }
            logParameters(path.node);
        },
        ClassPrivateMethod: function (path) {
            // path.node.key has type PrivateName
            logIdentifierJSON("Method", path.node.key.id.name, locationString(path.node.key.id));
            logParameters(path.node);
        },
        ClassProperty: function (path) {
            if (path.node.key.type === "Identifier") {
                logIdentifierJSON("Property", path.node.key.name, locationString(path.node.key));
            }
        },
        ClassPrivateProperty: function (path) {
            // path.node.key is PrivateName
            logIdentifierJSON("Property", path.node.key.id.name, locationString(path.node.key.id));
        },
        MemberExpression: function (path) {
            // Should be either Identifier or PrivateName if static (a.b) property
            // else Expression if computed (a[b]) property
            let identifier = null;
            if (path.node.property.type === "Identifier") {
                identifier = path.node.property;
            } else if (path.node.property.type === "PrivateName") {
                identifier = path.node.property.id;
            }
            if (identifier !== null) {
                logIdentifierJSON("Member", identifier.name, locationString(identifier));
            }
        },
        CatchClause: function(path) {
            logParameter(path.node.param);
        },
        LabeledStatement: function (path) {
            logIdentifierJSON("StatementLabel", path.node.label.name, locationString(path.node.label));
        },
        VariableDeclarator: function (path) {
            logIdentifierJSON("Variable", path.node.id.name, locationString(path.node.id));
        },
        //Identifier: function (path) {
        //    const loc = locationString(path.node);
        //    logIdentifierJSON("Unrecognised", path.node.name, loc);
        //},
        StringLiteral: function (path) {
            const loc = locationString(path.node);
            logLiteralJSON("String", path.node.value, loc, false, path.node.extra);
        },
        NumericLiteral: function (path) {
            const loc = locationString(path.node);
            logLiteralJSON("Numeric", path.node.value, loc, false, path.node.extra);
        },
        ArrayExpression: function (path) {
            path.traverse(arrayVisitor);
            path.skip();
        },
        TemplateElement: function (path) {
            const loc = locationString(path.node);
            logLiteralJSON("StringTemplate", path.node.value.raw, loc, false, path.node.value);
        }
    };
    traverse(ast, astVisitor);
}

function findLiteralsAndIdentifiers(source, printDebug) {
    const ast = parser.parse(source);

    // walk the AST and print out any literals
    if (printDebug) {
        console.log(JSON.stringify(ast, null, "  "));
    }

    traverseAst(ast);

    const allJson = "[\n" + parseOutputLines.join(",\n") + "\n]";
    console.log(allJson);
}

function main() {
    const printDebug = false;
    // https://github.com/nodejs/help/issues/2663
    // Referencing process.stdin.fd (actually just process.stdin) causes stdin to become nonblocking
    // Therefore running this in a terminal in interactive mode with no file piped into stdin will
    // cause the read to fail with EAGAIN
    // Passing the raw '0' as the fd avoids this issue.
    const sourceFile = process.argv.length >= 3 ? process.argv[2] : 0;
    const sourceCode = fs.readFileSync(sourceFile, "utf8");
    if (printDebug) {
        console.log("Read source:");
        console.log(sourceCode);
    }
    findLiteralsAndIdentifiers(sourceCode, printDebug);
}

main();
