// eslint no-var: 0
// jshint esversion: 6
// jshint node: true
"use strict";

import fs from "fs";
import parser from "@babel/parser";

// See https://github.com/babel/babel/issues/13855
import _traverse from "@babel/traverse";

const traverse = _traverse.default;

// used to signal to parent process that parsing could not complete due to syntax errors
const fatalSyntaxErrorMarker = "FATAL SYNTAX ERROR";

function position(node) {
    return (node.loc !== null) ? [node.loc.start.line,node.loc.start.column] : [];
}

// Holds all parsing data for a single file
class ParseData {
    constructor() {
        // holds token information (function, variable names)
        this.tokens = [];
        // holds status information (info, errors)
        this.status = [];
    }

    static makeOutputDict(type, subtype, data, pos, extra = null) {
        return { "type": type, "subtype": subtype, data: data, pos: pos, extra: (extra === null) ? {} : extra };
    }

    logError(errorType, message, pos) {
        this.status.push(ParseData.makeOutputDict("Error", errorType, message, pos));
    }

    logInfo(infoType, message) {
        this.status.push(ParseData.makeOutputDict("Info", infoType, message, []));
    }

    logComment(commentType, comment, pos) {
        this.tokens.push(ParseData.makeOutputDict("Comment", commentType, comment, pos));
    }

    logIdentifierOrPrivateName(identifierType, node) {
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
        let pos = position(identifierNode);

        if (identifierNode.name === undefined) {
            console.log("Error: undefined identifier name at pos " + pos);
        }

        this.tokens.push(ParseData.makeOutputDict("Identifier", identifierType, name, pos));
    }

    logLiteral(literalType, value, pos, inArray, extra = null) {
        if (value === undefined) {
            console.log("Error: undefined literal value at pos " + pos);
            return;
        }

        if (extra === null) {
            extra = {};
        }

        extra.array = inArray;
        this.tokens.push(ParseData.makeOutputDict("Literal", literalType, value, pos, extra));
    }
}

function visitIdentifierOrPrivateName(path, parseData) {
    const node = path.node;
    const parentNode = path.parentPath.node;
    const parentParentNode = path.parentPath.parentPath.node;

    switch (parentNode.type) {
        case "ObjectProperty":
            if (node === parentNode.key) {
                parseData.logIdentifierOrPrivateName("Variable", node);
            }
            break;
        case "ArrayPattern":
            parseData.logIdentifierOrPrivateName("Variable", node);
            break;
        case "VariableDeclarator":
            if (node === parentNode.id) {
                parseData.logIdentifierOrPrivateName("Variable", node);
            }
            break;
        case "FunctionDeclaration":
            if (node === parentNode.id) {
                parseData.logIdentifierOrPrivateName("Function", node);
            } else {
                parseData.logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "LabeledStatement":
            parseData.logIdentifierOrPrivateName("StatementLabel", node);
            break;
        case "PrivateName":
            // processed already
            break;
        case "MemberExpression":
            if (node === parentNode.property) {
                parseData.logIdentifierOrPrivateName("Member", node);
            }
            break;
        case "CatchClause":
            parseData.logIdentifierOrPrivateName("Parameter", node);
            break;
        case "ClassPrivateMethod":
            // fall through
        case "ClassMethod":
            if (node === parentNode.key) {
                if (parentNode.kind !== "constructor") {
                    parseData.logIdentifierOrPrivateName("Method", node);
                }
            } else {
                parseData.logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "ClassPrivateProperty":
            // fall through
        case "ClassProperty":
            parseData.logIdentifierOrPrivateName("Property", node);
            break;
        case "AssignmentPattern":
            if (node === parentNode.left && parentParentNode.type === "FunctionDeclaration") {
                // function parameter with default value
                parseData.logIdentifierOrPrivateName("Parameter", node);
            }
            break;
        case "AssignmentExpression":
            if (node === parentNode.left) {
                parseData.logIdentifierOrPrivateName("AssignmentTarget", node);
            }
            break;
        case "ClassExpression":
            parseData.logIdentifierOrPrivateName("Class", node);
            break;
    }
}

/*
 disableScope prevents tracking of parsing context during traversal.
 In particular, this redeclared variables from crashing the traversal
 when the AST was produced from parsing with errorRecovery: true.
 */
function traverseAst(ast, parseData, disableScope) {
    /*
      TODO
       1. Consider adding state to allow distinction between elements from different arrays
       2. Consider logging names of decorators
     */
    const arrayVisitor = {
        noScope: disableScope,
        StringLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("String", path.node.value, loc, true, path.node.extra);
        },
        NumericLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Numeric", path.node.value, loc, true, path.node.extra);
        },
        BigIntLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Numeric", path.node.value, loc, true, path.node.extra);
        },
        RegExpLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Regexp", path.node.pattern, loc, true, path.node.extra);
        },
        TemplateElement: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("StringTemplate", path.node.value.raw, loc, true, path.node.value);
        }
    };

    const astVisitor = {
        noScope: disableScope,
        Identifier: function (path) {
            visitIdentifierOrPrivateName(path, this.parseData);
        },
        PrivateName: function (path) {
            visitIdentifierOrPrivateName(path, this.parseData);
        },
        StringLiteral: function (path) {
            const loc = position(path.node);
            this.parseData.logLiteral("String", path.node.value, loc, false, path.node.extra);
        },
        DirectiveLiteral: function (path) {
            // same as string literal
            const loc = position(path.node);
            this.parseData.logLiteral("String", path.node.value, loc, false, path.node.extra);
        },
        NumericLiteral: function (path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Numeric", path.node.value, loc, false, path.node.extra);
        },
        BigIntLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Numeric", path.node.value, loc, false, path.node.extra);
        },
        RegExpLiteral: function(path) {
            const loc = position(path.node);
            this.parseData.logLiteral("Regexp", path.node.pattern, loc, false, path.node.extra);
        },
        ArrayExpression: function (path) {
            path.traverse(arrayVisitor, { parseData });
            path.skip();
        },
        TemplateElement: function (path) {
            const loc = position(path.node);
            this.parseData.logLiteral("StringTemplate", path.node.value.raw, loc, false, path.node.value);
        }
    };

    traverse(ast, astVisitor, null, { parseData });
}

function parseFile(fileName, allowSyntaxErrors, includeAST) {
    const sourceCode = fs.readFileSync(fileName, "utf8");

    const parseData = new ParseData();
    parseData.logInfo("InputLength", sourceCode.length.toString());

    try {
        const ast = parser.parse(sourceCode, {
            errorRecovery: allowSyntaxErrors,
            sourceType: "unambiguous" // parser is allowed to parse input as either script or module
        });

        if (includeAST) {
            parseData.ast = ast;
        }

        for (let e of ast.errors) {
            let pos = `[${e.loc.line},${e.loc.column}]`;
            parseData.logError(e.name, `${e.code}: ${e.reasonCode}`, pos);
        }

        for (let c of ast.comments) {
            let loc = position(c);
            parseData.logComment(c.type, c.value, loc);
        }

        traverseAst(ast, parseData, allowSyntaxErrors);

    } catch (e) {
        if (e instanceof SyntaxError) {
            let pos = [e.loc.line, e.loc.column];
            parseData.logError(e.name, `${e.code}: ${e.reasonCode}`, pos);
            parseData.logError(e.name, `${fatalSyntaxErrorMarker} (unable to parse remainder of file)`, pos);
        } else {
            throw(e);
        }
    }

    return parseData;
}

function main() {
    /*
     If allowSyntaxErrors is false, any syntax error results in immediate termination
     of parsing for a file. If true, the parser will recover from minor syntax errors
     where possible. All error details are recorded in the returned parseData object.
     */
    let allowSyntaxErrors = false;

    let printDebug = false;

    /*
     Use stdin if no filename is specified.

     Referencing process.stdin.fd (actually just process.stdin) causes stdin to become nonblocking
     Therefore running this in a terminal in interactive mode with no file piped into stdin will
     cause the read to fail with EAGAIN. Passing the raw '0' as the fd avoids this issue.
     See https://github.com/nodejs/help/issues/2663
     */
    const sourceFile = process.argv.length >= 3 ? process.argv[2] : 0;

    if (process.argv[process.argv.length - 1] === "debug") {
        printDebug = true;
    }

    const parseData = parseFile(sourceFile, allowSyntaxErrors, printDebug);

    console.log(JSON.stringify(parseData, null, "  "));
}

main();
