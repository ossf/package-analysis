#!/usr/bin/env python3

"""
Custom tool to pretty-print JSON with certain fields compacted

Adapted from source of `python -m json.tool`
reference: github.com/python/cpython/blob/main/Lib/json/tool.py
"""

import json
import re
import sys


# Changes JSON structs that are formatted like:
#     {
#         "key1": ...
#     }
# into ones like
#     { "key1": ... }
struct_single_key_substitution = (
    re.compile('{$\\n^\\s*"(.+)": ?(.*)$\\n^\\s*}', re.MULTILINE),
    '{ "\\1": \\2 }'
)

# Changes JSON structs that are formatted like:
#     {
#         "key1": ...,
#         "key2": ...
#     }
# into ones like
#     { "key1": ..., "key2": ... }
struct_pair_substitution = (
    re.compile('{$\\n^\\s*"(.+)": ?(.*),$\\n^\\s*"(.+)": ?(.*)$\\n^\\s*}', re.MULTILINE),
    '{ "\\1": \\2, "\\3": \\4 }'
)

# Changes JSON structs that are formatted like:
#     {
#         "key1": ...,
#         "key2": ...,
#         "key3": ...
#     }
# into ones like
#     { "key1": ..., "key2": ..., "key3": ... }
struct_triple_substitution = (
    re.compile('{$\\n^\\s*"(.+)": ?(.*),$\\n^\\s*"(.+)": ?(.*),$\\n^\\s*"(.+)": ?(.*)$\\n^\\s*}', re.MULTILINE),
    '{ "\\1": \\2, "\\3": \\4, "\\5": \\6 }'
)

all_substitutions = (struct_single_key_substitution, struct_pair_substitution, struct_triple_substitution)


# Pretty prints a JSON object with newlines and indentation, then applies
# the substitutions above while maintaining indentation level.
def format_json(json_object) -> str:
    # pretty print with newlines and indent with 4 spaces,
    pretty_printed = json.dumps(json_object, indent=4)

    # apply all replacements in sequence
    for (pattern, replacement) in all_substitutions:
        pretty_printed = re.sub(pattern, replacement, pretty_printed)

    return pretty_printed


def main(args: list[str]):
    if "--help" in args:
        print(f"Usage: {args[0]} [<infile> [<outfile>]]")
        return

    input_path = args[1] if len(args) >= 2 else None
    output_path = args[2] if len(args) >= 3 else None

    if input_path:
        with open(input_path) as infile:
            json_object = json.load(infile)
    else:
        json_object = json.load(sys.stdin)

    custom_formatted_json = format_json(json_object)

    if output_path:
        with open(output_path, "w", encoding="utf-8") as outfile:
            outfile.write(custom_formatted_json)
            outfile.write("\n")
    else:
        print(custom_formatted_json)


if __name__ == '__main__':
    try:
        main(sys.argv)
    except BrokenPipeError as exc:
        sys.exit(exc.errno)
    except ValueError as e:
        raise SystemExit(e)

