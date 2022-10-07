#!/usr/bin/env node
/* eslint no-var: 0 */

const parser = require("@babel/parser");
const fs = require("fs");

function locationString(node) {
    return (node.loc != null) ? `[${node.loc.start.line},${node.loc.start.column}]` : "[]"
}

const parseOutputLines = []

function logJSON(type, subtype, name, pos, array, extra = null) {
    const extraValue = (extra !== null) ? `${JSON.stringify(extra)}` : "{}"
    const arrayValue = (array !== null) ? array : false
    const json = `{"type":"${type}","subtype":"${subtype}","data":${JSON.stringify(name)},"pos":${pos},` +
        `"array":${arrayValue}, "extra":${extraValue}}`
    parseOutputLines.push(json)
}
function logIdentifierJSON(subtype, name, pos, extra = null) {
    logJSON("Identifier", subtype, name, pos, null, extra)
}

function logLiteralJSON(subtype, value, pos, inArray, extra = null) {
    logJSON("Literal", subtype, value, pos, inArray, extra)
}

function logParameters(node) {
    const n = node
    for (let i = 0; i < n.params.length; i++) {
        p = n.params[i]
        switch (p.type) {
            // simple function parameter name:
            // function f(x)
            case "Identifier":
                logIdentifierJSON("Parameter", p.name, locationString(p))
                break
            // parameter with default value:
            // function f(x = 3)
            case "AssignmentPattern":
                if (p.left.type === "Identifier") {
                    logIdentifierJSON("Parameter", p.left.name, locationString(p))
                } else {
                    // This is too hard...
                    walkAst(p.left)
                }
                walkAst(p.right)
                break
        }
    }

}

function multiWalkAst(nodeList, isInArray = false) {
    if (nodeList !== null) {
        for (let i = 0; i < nodeList.length; i++) {
            walkAst(nodeList[i], isInArray)
        }
    }
}

function walkAst(startNode, isInArray = false, printDebug = false) {
    // walk the AST and print out any literals
    if (startNode == null) {
        return
    }
    const n = startNode;

    if (printDebug) {
        console.log(`# type: ${n.type}`)
    }

    const loc = locationString(n)
    switch (n.type) {
        case "File":
            walkAst(n.program)
            break
        case "Program":
        // fall-through
        case "BlockStatement":
            multiWalkAst(n.body, isInArray)
            break
        case "ArrowFunctionExpression":
        // fall-through
        case "FunctionDeclaration":
            if (n.id !== null) {
                logIdentifierJSON("Function", n.id.name, locationString(n.id))
            }
            logParameters(n)
            walkAst(n.body)
            break
        case "AwaitExpression":
            walkAst(n.argument)
            break
        case "AssignmentExpression":
            if (n.left.type === "Identifier") {
                logIdentifierJSON("AssignmentTarget", n.left.name, locationString(n.left))
            } else {
                walkAst(n.left)
            }
            walkAst(n.right)
            break
        case "BinaryExpression":
            walkAst(n.left)
            walkAst(n.right)
            break
        case "ExpressionStatement":
            walkAst(n.expression)
            break
        case "CallExpression":
            walkAst(n.callee)
            multiWalkAst(n.arguments, isInArray)
            break
        case "ClassExpression":
            if (n.id !== null) {
                logIdentifierJSON("Class", n.id.name, locationString(n.id))
            }
            walkAst(n.body)
            // TODO superclass, decorators?
            break
        case "ClassBody":
            multiWalkAst(n.body)
            break
        case "ClassMethod":
            if (n.kind !== "Constructor" && n.key.type === "Identifier") {
                logIdentifierJSON("Method", n.key.name, locationString(n.key))
            }
            logParameters(n)
            walkAst(n.body)
            break
        case "ClassPrivateMethod":
            // key is PrivateName
            logIdentifierJSON("Method", n.key.id.name, locationString(n.key.id))
            logParameters(n)
            walkAst(n.body)
            break
        case "ClassProperty":
            if (n.key.type === "Identifier") {
                logIdentifierJSON("Property", n.key.name, locationString(n.key))
            } else {
                walkAst(n.key)
            }
            walkAst(n.value)
            break
        case "ClassPrivateProperty":
            // key is PrivateName
            logIdentifierJSON("Property", n.key.id.name, locationString(n.key.id))
            walkAst(n.value)
            break
        case "StaticBlock":
            multiWalkAst(n.body)
            break
        case "MemberExpression":
            walkAst(n.object)
            // Should be either Identifier or PrivateName if static (a.b) property
            // else Expression if computed (a[b]) property
            switch (n.property.type) {
                case "Identifier":
                    logIdentifierJSON("Member", n.property.name, locationString(n.property))
                    break
                case "PrivateName":
                    logIdentifierJSON("Member", n.property.id.name, locationString(n.property.id))
                    break
                default:
                    if (!n.computed && printDebug) {
                        console.log(`Warning: MemberExpression had property of type ${n.property.type} but was not computed`)
                        console.log(n.property)
                    }
                    walkAst(n.property)
                    break
            }
            break
        case "ThisExpression":
            // nothing to do
            break
        case "ReturnStatement":
            walkAst(n.argument)
            break
        case "LabeledStatement":
            logIdentifierJSON("StatementLabel", n.label.name, locationString(n.label))
            walkAst(n.body)
            break
        case "BreakStatement":
            // fall-through
        case "ContinueStatement":
            // the only thing of interest would be the label, which must already have been defined somewhere else
            break
        case "WhileStatement":
            // fall-through
        case "DoWhileStatement":
            walkAst(n.test)
            walkAst(n.body)
            break
        case "ForStatement":
            walkAst(n.init)
            walkAst(n.test)
            walkAst(n.update)
            walkAst(n.body)
            break
        case "ForOfStatement":
            // fall-through
        case "ForInStatement":
            walkAst(n.left)
            walkAst(n.right)
            walkAst(n.body)
            break
        case "TryStatement":
            walkAst(n.block)
            walkAst(n.handler)
            walkAst(n.finalizer)
            break
        case "ThrowStatement":
            walkAst(n.argument)
            break
        case "CatchClause":
            walkAst(n.param)
            walkAst(n.body)
            break
        case "IfStatement":
            walkAst(n.test)
            walkAst(n.consequent)
            walkAst(n.alternate)
            break
        case "SwitchStatement":
            walkAst(n.discriminant)
            multiWalkAst(n.cases)
            break
        case "SwitchCase":
            walkAst(n.test)
            multiWalkAst(n.consequent)
            break
        case "VariableDeclaration":
            multiWalkAst(n.declarations, isInArray)
            break
        case "VariableDeclarator":
            logIdentifierJSON("Variable", n.id.name, locationString(n.id))
            walkAst(n.init)
            break
        case "Identifier":
            logIdentifierJSON("Other", n.name, loc)
            break
        case "StringLiteral":
            logLiteralJSON("String", n.value, loc, isInArray, n.extra)
            break
        case "NumericLiteral":
            logLiteralJSON("Numeric", n.value, loc, isInArray, n.extra)
            break
        case "ArrayExpression":
            multiWalkAst(n.elements, true)
            break
        case "TemplateLiteral":
            multiWalkAst(n.quasis, isInArray)
            multiWalkAst(n.expressions, isInArray)
            break
        case "TemplateElement":
            logLiteralJSON("StringTemplate", n.value.raw, loc, isInArray, n.value)
            break
        default:
            if (printDebug) {
                console.log(`Found unknown node of type ${n.type} node @ ${loc}`);
                console.log(n)
            }
    }
}



function findLiteralsAndIdentifiers(source, printDebug) {
    const ast = parser.parse(source);

    // walk the AST and print out any literals
    if (printDebug) {
        console.log(JSON.stringify(ast, null, "  "));
    }

    walkAst(ast, printDebug)
    allJson = "[\n" + parseOutputLines.join(",\n") + "\n]"
    console.log(allJson)
}

function main() {
    const printDebug = false
    // https://github.com/nodejs/help/issues/2663
    // Referencing process.stdin.fd (actually just process.stdin) causes stdin to become nonblocking
    // Therefore running this in a terminal in interactive mode with no file piped into stdin will
    // cause the read to fail with EAGAIN
    // Passing the raw '0' as the fd avoids this issue.
    const sourceFile = process.argv.length >= 3 ? process.argv[2] : 0; // process.stdin.fd;
    const sourceCode = fs.readFileSync(sourceFile, "utf8");
    if (printDebug) {
        console.log("Read source:")
        console.log(sourceCode)
    }
    findLiteralsAndIdentifiers(sourceCode, printDebug)
}

main()
