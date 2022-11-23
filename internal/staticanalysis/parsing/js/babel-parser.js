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

function logIdentifierOrPrivateName(identifierType, node) {
    // if node is a PrivateName, the corresponding Identifier can be found as node.id
    let identifierNode;
    switch (node.type) {
        case "Identifier":
            identifierNode = node;
            break;
        case "PrivateName":
            identifierNode = node.id;
            break;
        default:
            console.log("Error: logIdentifierNodeJSON passed a node of type " + node.type);
            return;
    }

    let name = identifierNode.name;
    let pos = locationString(identifierNode)

    if (identifierNode.name === undefined) {
        console.log("Error: undefined identifier name at pos " + pos);
    }

    logJSON("Identifier", identifierType, name, pos, null, null);
}

function logLiteral(literalType, value, pos, inArray, extra = null) {
    if (value === undefined) {
        console.log("Error: undefined literal value at pos " + pos);
        return;
    }
    logJSON("Literal", literalType, value, pos, inArray, extra);
}

function visitIdentifierOrPrivateName(path) {
    const node = path.node;
    const parentNode = path.parentPath.node;
    const parentParentNode = path.parentPath.parentPath.node;

    switch (parentNode.type) {
        case "ObjectProperty":
        // fall through
            if (node === parentNode.key) {
                logIdentifierOrPrivateName("Variable", node);
            }
            break;
        case "ArrayPattern":
            logIdentifierOrPrivateName("Variable", node);
            break;
        case "VariableDeclarator":
            if (node === parentNode.id) {
                logIdentifierOrPrivateName("Variable", node);
            }
            break;
        case "FunctionDeclaration":
            if (node === parentNode.id) {
                logIdentifierOrPrivateName("Function", node);
            } else {
                logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "LabeledStatement":
            logIdentifierOrPrivateName("StatementLabel", node);
            break;
        case "PrivateName":
            // processed already
            break;
        case "MemberExpression":
            if (node === parentNode.property) {
                logIdentifierOrPrivateName("Member", node);
            }
            break;
        case "CatchClause":
            logIdentifierOrPrivateName("Parameter", node);
            break;
        case "ClassPrivateMethod":
            // fall through
        case "ClassMethod":
            if (node === parentNode.key) {
                if (parentNode.kind !== "constructor") {
                    logIdentifierOrPrivateName("Method", node);
                }
            } else {
                logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "ClassPrivateProperty":
            // fall through
        case "ClassProperty":
            logIdentifierOrPrivateName("Property", node);
            break;
        case "AssignmentPattern":
            if (node === parentNode.left && parentParentNode.type === "FunctionDeclaration") {
                // function parameter with default value
                logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "AssignmentExpression":
            if (node === parentNode.left) {
                logIdentifierOrPrivateName("AssignmentTarget", node);
            }
            break;
        case "ClassExpression":
            logIdentifierOrPrivateName("Class", node);
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
            logLiteral("String", path.node.value, loc, true, path.node.extra);
        },
        NumericLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Numeric", path.node.value, loc, true, path.node.extra);
        },
        TemplateElement: function(path) {
            const loc = locationString(path.node);
            logLiteral("StringTemplate", path.node.value.raw, loc, true, path.node.value);
        }
    };

    const astVisitor = {
        Identifier: visitIdentifierOrPrivateName,
        PrivateName: visitIdentifierOrPrivateName,
        StringLiteral: function (path) {
            const loc = locationString(path.node);
            logLiteral("String", path.node.value, loc, false, path.node.extra);
        },
        DirectiveLiteral: function (path) {
            // same as string literal
            const loc = locationString(path.node);
            logLiteral("String", path.node.value, loc, false, path.node.extra);

        },
        NumericLiteral: function (path) {
            const loc = locationString(path.node);
            logLiteral("Numeric", path.node.value, loc, false, path.node.extra);
        },
        ArrayExpression: function (path) {
            path.traverse(arrayVisitor);
            path.skip();
        },
        TemplateElement: function (path) {
            const loc = locationString(path.node);
            logLiteral("StringTemplate", path.node.value.raw, loc, false, path.node.value);
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
