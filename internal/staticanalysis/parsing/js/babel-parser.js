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
    if (name === undefined) {
        console.log("Error: undefined identifier name at pos " + pos);
        return;
    }
    logJSON("Identifier", subtype, name, pos, null, extra);
}

// node is an Identifier node or PrivateName (which has a key attribute holding an Identifier)
function logIdentifierNodeJSON(subtype, node) {
    switch (node.type) {
        case "Identifier":
            logIdentifierJSON(subtype, node.name, locationString(node));
            break;
        case "PrivateName":
            logIdentifierJSON(subtype, node.id.name, locationString(node.id));
            break;
        default:
            console.log("Error: logIdentifierNodeJSON passed a node of type " + node.type);
            break;
    }
}

function logLiteralJSON(subtype, value, pos, inArray, extra = null) {
    if (value === undefined) {
        console.log("Error: undefined literal value at pos " + pos);
        return;
    }
    logJSON("Literal", subtype, value, pos, inArray, extra);
}

function visitIdentifierOrPrivateName(path) {
    const node = path.node;
    const parentNode = path.parentPath.node;
    const parentParentNode = path.parentPath.parentPath.node;

    switch (parentNode.type) {
        case "ObjectProperty":
        // fall through
            if (node === parentNode.key) {
                logIdentifierNodeJSON("Variable", node);
            }
            break;
        case "ArrayPattern":
            logIdentifierNodeJSON("Variable", node);
            break;
        case "VariableDeclarator":
            if (node === parentNode.id) {
                logIdentifierNodeJSON("Variable", node);
            }
            break;
        case "FunctionDeclaration":
            if (node === parentNode.id) {
                logIdentifierNodeJSON("Function", node);
            } else {
                logIdentifierNodeJSON("Parameter", node);
            }
            break;
        case "LabeledStatement":
            logIdentifierNodeJSON("StatementLabel", node);
            break;
        case "PrivateName":
            // processed already
            break;
        case "MemberExpression":
            if (node === parentNode.property) {
                logIdentifierNodeJSON("Member", node);
            }
            break;
        case "CatchClause":
            logIdentifierNodeJSON("Parameter", node);
            break;
        case "ClassPrivateMethod":
            // fall through
        case "ClassMethod":
            if (node === parentNode.key) {
                if (parentNode.kind !== "constructor") {
                    logIdentifierNodeJSON("Method", node);
                }
            } else {
                logIdentifierNodeJSON("Parameter", node);
            }
            break;
        case "ClassPrivateProperty":
            // fall through
        case "ClassProperty":
            logIdentifierNodeJSON("Property", node);
            break;
        case "AssignmentPattern":
            if (node === parentNode.left && parentParentNode.type === "FunctionDeclaration") {
                // function parameter with default value
                logIdentifierNodeJSON("Parameter", node);
            }
            break;
        case "AssignmentExpression":
            if (node === parentNode.left) {
                logIdentifierNodeJSON("AssignmentTarget", node);
            }
            break;
        case "ClassExpression":
            logIdentifierNodeJSON("Class", node);
            break;
    }
}

function traverseAst(ast) {
    /*
      TODO
       1. Consider adding state to allow distinction between elements from different arrays
       2. Consider logging names of decorators
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
        Identifier: visitIdentifierOrPrivateName,
        PrivateName: visitIdentifierOrPrivateName,
        StringLiteral: function (path) {
            const loc = locationString(path.node);
            logLiteralJSON("String", path.node.value, loc, false, path.node.extra);
        },
        DirectiveLiteral: function (path) {
            // same as string literal
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
    const syntaxErrorExitCode = 33;
    let printDebug = false;
    // https://github.com/nodejs/help/issues/2663
    // Referencing process.stdin.fd (actually just process.stdin) causes stdin to become nonblocking
    // Therefore running this in a terminal in interactive mode with no file piped into stdin will
    // cause the read to fail with EAGAIN
    // Passing the raw '0' as the fd avoids this issue.
    const sourceFile = process.argv.length >= 3 ? process.argv[2] : 0;
    if (process.argv[process.argv.length - 1] === "debug") {
        printDebug = true;
    }

    const sourceCode = fs.readFileSync(sourceFile, "utf8");
    if (printDebug) {
        console.log("Read source:");
        console.log(sourceCode);
    }
    try {
        findLiteralsAndIdentifiers(sourceCode, printDebug);
    } catch (e) {
        if (e instanceof SyntaxError) {
            process.exit(syntaxErrorExitCode);
        } else {
            throw(e);
        }
    }
}

main();
