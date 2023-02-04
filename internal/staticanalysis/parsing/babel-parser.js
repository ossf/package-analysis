// eslint no-var: 0
// jshint esversion: 6
// jshint node: true
"use strict";

import fs from "fs";
import parser from "@babel/parser";

// See https://github.com/babel/babel/issues/13855
import _traverse from "@babel/traverse";

const traverse = _traverse.default;

// If the parser encounters an unrecoverable syntax error (which may indicate
// that the input is not JavaScript, the program will exit with this exit code.
const SYNTAX_ERROR_EXIT_CODE = 33;

function locationString(node) {
    return (node.loc !== null) ? `[${node.loc.start.line},${node.loc.start.column}]` : "[]";
}

const parseOutputLines = [];

function logJSON(type, subtype, data, pos, extra = null) {
    const extraValue = (extra !== null) ? `${JSON.stringify(extra)}` : "{}";
    const json = `{"type":"${type}","subtype":"${subtype}","data":${JSON.stringify(data)},` +
        `"pos":${pos},"extra":${extraValue}}`;
    parseOutputLines.push(json);
}

function logError(errorType, message, pos) {
    logJSON("Error", errorType, message, pos);
}

function logInfo(infoType, message) {
    logJSON("Info", infoType, message, "[]");
}

function logComment(commentType, comment, pos) {
    logJSON("Comment", commentType, comment, pos);
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
    let pos = locationString(identifierNode);

    if (identifierNode.name === undefined) {
        console.log("Error: undefined identifier name at pos " + pos);
    }

    logJSON("Identifier", identifierType, name, pos);
}

function logLiteral(literalType, value, pos, inArray, extra = null) {
    if (value === undefined) {
        console.log("Error: undefined literal value at pos " + pos);
        return;
    }

    if (extra === null) {
        extra = {};
    }

    extra.array = inArray;
    logJSON("Literal", literalType, value, pos, extra);
}

function visitIdentifierOrPrivateName(path) {
    const node = path.node;
    const parentNode = path.parentPath.node;
    const parentParentNode = path.parentPath.parentPath.node;

    switch (parentNode.type) {
        case "ObjectProperty":
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

/*
 disableScope prevents tracking of parsing context during traversal.
 In particular, this redeclared variables from crashing the traversal
 when the AST was produced from parsing with errorRecovery: true.
 */
function traverseAst(ast, disableScope) {
    /*
      TODO
       1. Consider adding state to allow distinction between elements from different arrays
       2. Consider logging names of decorators
     */
    const arrayVisitor = {
        noScope: disableScope,
        StringLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("String", path.node.value, loc, true, path.node.extra);
        },
        NumericLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Numeric", path.node.value, loc, true, path.node.extra);
        },
        BigIntLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Numeric", path.node.value, loc, true, path.node.extra);
        },
        RegExpLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Regexp", path.node.pattern, loc, true, path.node.extra);
        },
        TemplateElement: function(path) {
            const loc = locationString(path.node);
            logLiteral("StringTemplate", path.node.value.raw, loc, true, path.node.value);
        }
    };

    const astVisitor = {
        noScope: disableScope,
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
        BigIntLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Numeric", path.node.value, loc, false, path.node.extra);
        },
        RegExpLiteral: function(path) {
            const loc = locationString(path.node);
            logLiteral("Regexp", path.node.pattern, loc, false, path.node.extra);
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

function main() {
    /*
     If false, syntax errors result in immediate termination of the program with
     SYNTAX_ERROR_EXIT_CODE, and JSON output will be suppressed.
     If true, the parser will recover from minor syntax errors where possible, and
     record error details in the output JSON. If not possible to recover, the program
     will still terminate with SYNTAX_ERROR_EXIT_CODE.
     */
    let allowSyntaxErrors = false;

    let printDebug = false;

    /*
     Referencing process.stdin.fd (actually just process.stdin) causes stdin to become nonblocking
     Therefore running this in a terminal in interactive mode with no file piped into stdin will
     cause the read to fail with EAGAIN. Passing the raw '0' as the fd avoids this issue.
     See https://github.com/nodejs/help/issues/2663
     */
    const sourceFile = process.argv.length >= 3 ? process.argv[2] : 0;
    if (process.argv[process.argv.length - 1] === "debug") {
        printDebug = true;
    }

    const sourceCode = fs.readFileSync(sourceFile, "utf8");
    logInfo("InputLength", sourceCode.length);

    let unrecoverableSyntaxError = false;

    try {
        const ast = parser.parse(sourceCode, {
            errorRecovery: allowSyntaxErrors,
            sourceType: "unambiguous" // parser is allowed to parse input as either script or module
        });

        if (printDebug) {
            console.log("AST");
            console.log(JSON.stringify(ast, null, "  "));
        }

        for (let e of ast.errors) {
            let pos = `[${e.loc.line},${e.loc.column}]`;
            logError(e.name, `${e.code}: ${e.reasonCode}`, pos);
        }

        for (let c of ast.comments) {
            let loc = locationString(c);
            logComment(c.type, c.value, loc);
        }

        traverseAst(ast, allowSyntaxErrors);

    } catch (e) {
        if (e instanceof SyntaxError) {
            unrecoverableSyntaxError = true;
            let pos = `[${e.loc.line},${e.loc.column}]`;
            logError(e.name, `${e.code}: ${e.reasonCode}`, pos);
        } else {
            throw(e);
        }
    }

    const allJSON = "[\n" + parseOutputLines.join(",\n") + "\n]";

    const suppressJSON = !allowSyntaxErrors && unrecoverableSyntaxError;
    if (!suppressJSON) {
        console.log(allJSON);
    }

    if (unrecoverableSyntaxError) {
        process.exit(SYNTAX_ERROR_EXIT_CODE);
    }
}

main();
